package jifa

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/rkonfj/opkit/pkg/config"
)

func Upload(filename string, fi io.Reader, typ string) error {

	uri := config.Conf.Jifa + "/jifa-api/file/upload"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("data", filename)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, fi)
	if err != nil {
		return err
	}

	_ = writer.WriteField("uploadName", "upload")
	_ = writer.WriteField("type", typ)

	err = writer.Close()
	if err != nil {
		return err
	}

	rpcctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(rpcctx, "POST", uri, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return nil
}
