
package line

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/png"
	"image/jpeg"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/url"
	"net/http"
	"net/textproto"
	"strings"
)

const (
	lineNotifyURL = "https://notify-api.line.me/api/notify"
)

func NotifyLineMessage(message, token string) error {
	if token == "" {
		return fmt.Errorf("line: no token")
	} else if message == "" {
		return fmt.Errorf("line: no message")
	}

	data := url.Values{"message": {message}}
	r, _ := http.NewRequest("POST", lineNotifyURL, strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("line: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("line: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("line: %s: %v", respBody, err)
	}
	return nil
}

func convert2Jpeg(img io.Reader) (io.Reader, error) {
	dimg, _, err := image.Decode(img)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	err = jpeg.Encode(&b, dimg, &jpeg.Options{Quality: 100})
	return &b, err
}

func NotifyLineImage(img io.Reader, filename, message, token string) error {
	if token == "" {
		return fmt.Errorf("line: no token")
	}

	body := &bytes.Buffer{}
	mw   := multipart.NewWriter(body)

	ffw, err := mw.CreateFormField("message")
	if err != nil {
		return fmt.Errorf("line: %v", err)
	}
	if _, err = ffw.Write([]byte(message)); err != nil {
		return fmt.Errorf("line: %v", err)
	}

	part := make(textproto.MIMEHeader)
	part.Set("Content-Disposition", `form-data; name="imageFile"; filename=`+filename)
	part.Set("Content-Type", "image/jpeg")

	fw, err := mw.CreatePart(part)
	if err != nil {
		return fmt.Errorf("line: %v", err)
	}

	cimg, err := convert2Jpeg(img)
	if err != nil {
		return fmt.Errorf("line: %v", err)
	}

	_, err = io.Copy(fw, cimg)
	if err != nil {
		return fmt.Errorf("line: %v", err)
	}
	err = mw.Close()
	if err != nil {
		return fmt.Errorf("line: %v", err)
	}

	r, _ := http.NewRequest("POST", lineNotifyURL, body)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {	
		return fmt.Errorf("line: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("line: %v", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("line: %s: %v", respBody, err)
	}
	return nil
}
