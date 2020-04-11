package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/prometheus/alertmanager/notify/webhook"
)

const lineNotifyTokenEnvKey = "LINE_NOTIFY_TOKEN"
const lineNotifyEndPoint = "https://notify-api.line.me/api/notify"

type amLineNotifyHandler struct {
	lineNotifyToken string
	lineNotifyTmpl  *template.Template
}

func newHandler(lineNotifyToken, lineNotifyTmplPath string) (*amLineNotifyHandler, error) {
	lineNotifyTmpl, err := template.ParseFiles(lineNotifyTmplPath)
	if err != nil {
		return nil, err
	}

	return &amLineNotifyHandler{
		lineNotifyToken: lineNotifyToken,
		lineNotifyTmpl:  lineNotifyTmpl,
	}, nil
}

func (h *amLineNotifyHandler) notify(message *webhook.Message) error {
	var msgBuf bytes.Buffer
	err := h.lineNotifyTmpl.Execute(&msgBuf, message)
	if err != nil {
		return err
	}

	values := url.Values{}
	values.Add("message", msgBuf.String())

	req, err := http.NewRequest("POST", lineNotifyEndPoint, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+h.lineNotifyToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("status code %d: %s", res.StatusCode, string(resBody))
	}

	return nil
}

func (h *amLineNotifyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		var message webhook.Message

		err := json.NewDecoder(req.Body).Decode(&message)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Fprintln(w, "OK")

		go func() {
			if err := h.notify(&message); err != nil {
				log.Println("error:", err)
			}
		}()

	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <addr> <template> [<lineNotifyToken>]\n", os.Args[0])
		os.Exit(1)
	}

	addr := os.Args[1]
	lineNotifyTmplPath := os.Args[2]

	var lineNotifyToken string
	if len(os.Args) >= 4 {
		lineNotifyToken = os.Args[3]
	} else {
		var ok bool
		lineNotifyToken, ok = os.LookupEnv(lineNotifyTokenEnvKey)
		if !ok {
			log.Fatalf("Neither the second argument nor %s is provided for LINE Notify token\n", lineNotifyTokenEnvKey)
		}
	}

	handler, err := newHandler(lineNotifyToken, lineNotifyTmplPath)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", handler)

	log.Println("serving on", addr)
	http.ListenAndServe(addr, nil)
}
