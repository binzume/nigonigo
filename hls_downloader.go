package nigonigo

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type HLSDownloader struct {
	HttpClient    *http.Client
	TargetGroupID string
	iv            []byte
	key           []byte
}

func NewHLSDownloader(client *http.Client) *HLSDownloader {
	if client == nil {
		client = &http.Client{}
	}
	return &HLSDownloader{HttpClient: client}
}

func (c *HLSDownloader) httpGet(ctx context.Context, url string) ([]byte, http.Header, error) {
	req, err := newGetReq(url, nil)
	if err != nil {
		return nil, nil, err
	}
	req = req.WithContext(ctx)
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	if res.StatusCode != 200 {
		return nil, nil, fmt.Errorf("invalid status code :%v", res.StatusCode)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	return body, res.Header, err
}

func (c *HLSDownloader) getBytes(ctx context.Context, url string) ([]byte, error) {
	body, _, err := c.httpGet(ctx, url)
	return body, err
}

func relative(base, relative string) string {
	u1, err := url.Parse(base)
	if err != nil {
		Logger.Fatal(err)
	}
	u2, err := url.Parse(relative)
	if err != nil {
		Logger.Fatal(err)
	}
	return u1.ResolveReference(u2).String()
}

func (c *HLSDownloader) GetSegment(ctx context.Context, url string, w io.Writer) error {
	req, err := newGetReq(url, nil)
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
	data, header, err := c.httpGet(ctx, url)
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

	// Remove padding
	if strings.Contains(header.Get("content-type"), "/mp4") {
		p := bytes.Index(out, []byte("mdat"))
		if p >= 4 {
			sz := (int(out[p-4]) << 24) | (int(out[p-3]) << 16) | (int(out[p-2]) << 8) | int(out[p-1])
			sz = p - 4 + sz
			if sz < len(out) && sz >= len(data)-cipherBlock.BlockSize() {
				out = out[:sz]
			}
		}
	}

	_, err = w.Write(out)
	return err
}

type hlsPlaylist struct {
	MediaSequence int
	SubPlaylists  []string
	Lines         []string
	Endlist       bool
	Medias        []map[string]string
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
			} else if tag[0] == "#EXT-X-MEDIA" && len(tag) >= 2 {
				m := map[string]string{}
				for _, s := range strings.Split(tag[1], ",") {
					kv := strings.SplitN(s, "=", 2)
					m[kv[0]] = strings.Trim(kv[1], "\"")
				}
				list.Medias = append(list.Medias, m)
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
	if c.TargetGroupID != "" && len(list.Medias) > 0 {
		for _, m := range list.Medias {
			if m["GROUP-ID"] == c.TargetGroupID {
				return c.downloadInternal(ctx, relative(url, m["URI"]), w, start)
			}
		}
	}
	if len(list.SubPlaylists) > 0 {
		// TODO: select better stream.
		return c.downloadInternal(ctx, relative(url, list.SubPlaylists[0]), w, start)
	}

	reqInterval := 100 * time.Millisecond
	re := regexp.MustCompile(`^#EXT-X-KEY:METHOD=AES-128,URI="([^"]+)",IV=0x([0-9a-fA-F]+)`)
	mapRe := regexp.MustCompile(`^#EXT-X-MAP:URI="([^"]+)"`)
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
			Logger.Println(keyURL)
			Logger.Println(ivHex)
			c.iv, _ = hex.DecodeString(ivHex)
			c.key, err = c.getBytes(ctx, keyURL)
			if len(c.key) != 16 {
				return fmt.Errorf("invalid key length : %v", len(c.key))
			}
			if err != nil {
				return err
			}
		}
		match = mapRe.FindStringSubmatch(line)
		if match != nil {
			err = c.GetSegment(ctx, relative(url, match[1]), w)
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
