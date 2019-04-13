package nigonigo

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
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

func (c *HLSDownloader) getBytes(url string) ([]byte, error) {
	return GetContent(c.HttpClient, url)
}

func (c *HLSDownloader) SaveKey(keyURL string) error {
	key, err := c.getBytes(keyURL)
	if err != nil {
		return err
	}
	w, _ := os.Create("./hls.key")
	_, err = w.Write(key)
	return nil
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

func (c *HLSDownloader) GetSegment(url string, w io.Writer) error {
	req, err := NewGetReq(url, nil)
	if err != nil {
		return err
	}
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, err = io.Copy(w, res.Body)
	return err
}

func (c *HLSDownloader) GetSegment2(url string, w io.Writer, key, iv []byte) error {
	// TODO io.Reader/Writer
	data, err := c.getBytes(url)
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

func (c *HLSDownloader) Downaload(ctx context.Context, url string, w io.Writer) error {
	res, err := c.getBytes(url)
	if err != nil {
		return err
	}
	playlist := string(res)

	re := regexp.MustCompile(`^#EXT-X-KEY:METHOD=AES-128,URI="([^"]+)",IV=0x([0-9a-fA-F]+)`)
	lines := strings.Split(playlist, "\n")
	for i, line := range lines {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			// TODO: select better stream.
			time.Sleep(1 * time.Second)
			return c.Downaload(ctx, relative(url, lines[i+1]), w)
		}
		match := re.FindStringSubmatch(line)
		if match != nil {
			keyURL := match[1]
			ivHex := match[2]
			log.Println(keyURL)
			log.Println(ivHex)
			c.iv, _ = hex.DecodeString(ivHex)
			c.key, err = c.getBytes(keyURL)
			if len(c.key) != 16 {
				return fmt.Errorf("invalid key length : %v", len(c.key))
			}
			if err != nil {
				return err
			}
		}
		if len(line) > 0 && !strings.HasPrefix(line, "#") {
			if c.key != nil {
				err = c.GetSegment2(relative(url, line), w, c.key, c.iv)
			} else {
				err = c.GetSegment(relative(url, line), w)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
