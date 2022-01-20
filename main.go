package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"
)

var wf *aw.Workflow
var TRANSLATE_URL = "http://api.fanyi.baidu.com/api/trans/vip/translate"

type TranslateResultElement struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

type TranslateResult struct {
	From    string                   `json:"from"`
	To      string                   `json:"to"`
	Results []TranslateResultElement `json:"trans_result"`
}

func calculateSign(appid string, key string, query string) (string, string) {
	salt := fmt.Sprintf("%d", time.Now().Unix())
	str1 := appid + query + salt + key
	hash := md5.Sum([]byte(str1))
	return hex.EncodeToString(hash[:]), salt
}

func translate(content string, appid string, key string, to string) (*TranslateResult, error) {
	client := http.DefaultClient
	req, err := http.NewRequest(http.MethodGet, TRANSLATE_URL, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	sign, salt := calculateSign(appid, key, content)
	q.Add("q", content)
	q.Add("appid", appid)
	q.Add("salt", salt)
	q.Add("sign", sign)
	q.Add("from", "auto")
	q.Add("to", to)
	req.URL.RawQuery = q.Encode()
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var result TranslateResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func run() {
	args := wf.Args()
	appid, _ := wf.Alfred.Env.Lookup("appid")
	key, _ := wf.Alfred.Env.Lookup("appkey")
	// 替换所有的换行符
	content := strings.Replace(args[0], "\n", " ", -1)

	r, err := translate(content, appid, key, args[1])

	if err != nil {
		wf.FatalError(err)
	}

	for _, result := range r.Results {

		params := url.Values{
			"query":  {result.Src},
			"right":  {result.Dst},
			"twoCol": {"1"},
		}

		quicklook :=
			"https://quicklook.maoyachen.com/?" + params.Encode()

		wf.NewItem(result.Dst).
			Subtitle(result.Src).
			Arg(result.Dst).
			Copytext(result.Dst).
			Quicklook(quicklook).
			Valid(true).
			Cmd().
			Arg(result.Dst).
			Subtitle("🔍Large Type")
	}

	wf.SendFeedback()

}

func main() {
	wf = aw.New()
	wf.Run(run)
}
