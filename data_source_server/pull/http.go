package pull

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func HttpRequest(method, rawUrl string, bodyMaps, headers map[string]string, timeout time.Duration) (result []byte, err error) {
	var (
		request  *http.Request
		response *http.Response
		res      []byte
	)
	if timeout <= 0 {
		timeout = 5
	}
	client := &http.Client{
		Timeout: timeout * time.Second,
	}

	// 请求的 body 内容
	data := url.Values{}
	for key, value := range bodyMaps {
		data.Set(key, value)
	}

	jsons := data.Encode()

	if request, err = http.NewRequest(method, rawUrl, strings.NewReader(jsons)); err != nil {
		return
	}

	if method == "GET" {
		request.URL.RawQuery = jsons
	}

	// 增加header头信息
	for key, val := range headers {
		request.Header.Set(key, val)
	}

	// 处理返回结果
	if response, err = client.Do(request); err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if res, err = io.ReadAll(response.Body); err != nil {
		return nil, err
	}
	return res, nil
}
