package nigonigo

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type HLSDownloader struct {
	HttpClient *http.Client
	iv         []byte
	key        []byte
}

func NewHLSDownloader(client *http.Client) *HLSDownloader {
	if client == nil {
		client = &http.Client{}
	}
	return &HLSDownloader{HttpClient: client}
}

func (c *HLSDownloader) getBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := NewGetReq(url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

// Simple hls client...

func relative(base, relative string) string {
	u1, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}
	u2, err := url.Parse(relative)
	if err != nil {
		log.Fatal(err)
	}
	return u1.ResolveReference(u2).String()
}

func (c *HLSDownloader) GetSegment(ctx context.Context, url string, w io.Writer) error {
	req, err := NewGetReq(url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, err = io.Copy(w, res.Body)
	return err
}

func (c *HLSDownloader) GetSegment2(ctx context.Context, url string, w io.Writer, key, iv []byte) error {
	// TODO io.Reader/Writer
	data, err := c.getBytes(ctx, url)
	if err != nil {
		return err
	}

	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	if len(data)%cipherBlock.BlockSize() != 0 {
		return fmt.Errorf("invalid length %v", len(data))
	}

	out := make([]byte, len(data))
	cbc := cipher.NewCBCDecrypter(cipherBlock, iv)
	cbc.CryptBlocks(out, data)

	_, err = w.Write(out)
	return err
}

type hlsPlaylist struct {
	MediaSequence int
	SubPlaylists  []string
	Lines         []string
	Endlist       bool
}

func parseHLSPlaylist(playlist []byte) *hlsPlaylist {
	lines := strings.Split(string(playlist), "\n")
	list := &hlsPlaylist{Lines: lines}
	for i, line := range lines {
		if strings.HasPrefix(line, "#") {
			tag := strings.SplitN(line, ":", 2)
			if tag[0] == "#EXT-X-STREAM-INF" {
				list.SubPlaylists = append(list.SubPlaylists, lines[i+1])
				lines[i+1] = "" // avoid download as .ts
			} else if tag[0] == "#EXT-X-MEDIA-SEQUENCE" && len(tag) >= 2 {
				list.MediaSequence, _ = strconv.Atoi(tag[1])
			} else if tag[0] == "#EXT-X-ENDLIST" {
				list.Endlist = true
			}
		}
	}
	return list
}

func (c *HLSDownloader) downloadInternal(ctx context.Context, url string, w io.Writer, start int) error {
	res, err := c.getBytes(ctx, url)
	if err != nil {
		return err
	}
	list := parseHLSPlaylist(res)
	if len(list.SubPlaylists) > 0 {
		// TODO: select better stream.
		return c.downloadInternal(ctx, relative(url, list.SubPlaylists[0]), w, start)
	}

	reqInterval := 100 * time.Millisecond
	re := regexp.MustCompile(`^#EXT-X-KEY:METHOD=AES-128,URI="([^"]+)",IV=0x([0-9a-fA-F]+)`)
	mediaSeq := list.MediaSequence
	for _, line := range list.Lines {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(reqInterval):
		}
		match := re.FindStringSubmatch(line)
		if match != nil {
			keyURL := match[1]
			ivHex := match[2]
			log.Println(keyURL)
			log.Println(ivHex)
			c.iv, _ = hex.DecodeString(ivHex)
			c.key, err = c.getBytes(ctx, keyURL)
			if len(c.key) != 16 {
				return fmt.Errorf("invalid key length : %v", len(c.key))
			}
			if err != nil {
				return err
			}
		}
		if len(line) > 0 && !strings.HasPrefix(line, "#") {
			mediaSeq++
			if mediaSeq <= start {
				continue
			}
			if c.key != nil {
				err = c.GetSegment2(ctx, relative(url, line), w, c.key, c.iv)
			} else {
				err = c.GetSegment(ctx, relative(url, line), w)
			}
			if err != nil {
				return err
			}
		}
	}
	if list.Endlist || mediaSeq <= start {
		return nil
	}
	// TODO sleep
	return c.downloadInternal(ctx, url, w, mediaSeq)
}

func (c *HLSDownloader) Download(ctx context.Context, url string, w io.Writer) error {
	return c.downloadInternal(ctx, url, w, -1)
}
