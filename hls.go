package nigonigo

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

func (c *Client) SaveKey(keyURL string) error {
	key, err := c.GetContent(keyURL)
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

func (c *Client) GetSegment(url string, w io.Writer) error {
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

func (c *Client) GetSegment2(url string, w io.Writer, key, iv []byte) error {
	// TODO io.Reader/Writer
	data, err := c.GetContent(url)
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

func (c *Client) StartHlsDownload(session *DMCSession, w io.Writer) error {
	time.Sleep(1 * time.Second)
	master, err := c.GetContent(session.ContentURI)
	if err != nil {
		return err
	}
	log.Println(string(master))

	// TODO: parse master.m3u8
	playlistURL := strings.Replace(session.ContentURI, "/master.m3u8", "/1/ts/playlist.m3u8", 1)
	if playlistURL == session.ContentURI {
		return errors.New("invalid url " + playlistURL)
	}

	res, err := c.GetContent(playlistURL)
	if err != nil {
		return err
	}
	playlist := string(res)

	re := regexp.MustCompile(`^#EXT-X-KEY:METHOD=AES-128,URI="([^"]+)",IV=0x([0-9a-fA-F]+)`)
	encrypted := false
	var iv []byte
	var key []byte
	for _, line := range strings.Split(playlist, "\n") {
		match := re.FindStringSubmatch(line)
		if match != nil {
			encrypted = true
			keyURL := match[1]
			ivHex := match[2]
			log.Println(keyURL)
			log.Println(ivHex)
			iv, _ = hex.DecodeString(ivHex)
			key, err = c.GetContent(keyURL)
			log.Println(len(key))
			if err != nil {
				return err
			}
		}
		if len(line) > 0 && !strings.HasPrefix(line, "#") {
			if encrypted {
				err = c.GetSegment2(relative(playlistURL, line), w, key, iv)
			} else {
				err = c.GetSegment(relative(playlistURL, line), w)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}