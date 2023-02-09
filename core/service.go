package core

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"wxcallback/lib/wxbizjsonmsgcrypt"
)

type MsgContent struct {
	ToUsername   string `json:"ToUserName,omitempty" xml:"ToUserName,omitempty"`
	FromUsername string `json:"FromUserName,omitempty" xml:"FromUserName,omitempty"`
	CreateTime   uint32 `json:"CreateTime,omitempty" xml:"CreateTime,omitempty"`
	MsgType      string `json:"MsgType,omitempty" xml:"MsgType,omitempty"`
	PicUrl       string `json:"PicUrl,omitempty" xml:"PicUrl,omitempty"`   // Image
	Content      string `json:"Content,omitempty" xml:"Content,omitempty"` // Text
	MediaId      string `json:"MediaId,omitempty" xml:"MediaId,omitempty"` // Image/Voice
	Format       string `json:"Format,omitempty" xml:"Format,omitempty"`   // Voice
	Msgid        uint64 `json:"MsgId,omitempty" xml:"MsgId,omitempty"`
	Agentid      uint32 `json:"AgentId,omitempty" xml:"AgentId,omitempty"`
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request, flag string, service *Service) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 Method Not Allowed"))
		return
	}
	urlValues := r.URL.Query()
	if service.VerifyUrl && urlValues.Get("echostr") != "" {
		s.handlerVerify(w, r, flag, service)
		return
	}
	wxCrypt := wxbizjsonmsgcrypt.NewWXBizMsgCrypt(service.Token, service.EncodingAesKey, service.AppID, wxbizjsonmsgcrypt.XMLType)
	var (
		reqMsgSign   = urlValues.Get("msg_signature")
		reqTimestamp = urlValues.Get("timestamp")
		reqNonce     = urlValues.Get("nonce")
	)
	buf := bytes.NewBuffer(nil)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		s.logger.Error(flag, fmt.Sprintf("read body fail: %s", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request"))
		return
	}
	reqData := buf.Bytes()
	msg, cryptErr := wxCrypt.DecryptMsg(reqMsgSign, reqTimestamp, reqNonce, reqData)
	if cryptErr != nil {
		s.logger.Error(flag, fmt.Sprintf("decrypt data fail: [%s] %s", strconv.Itoa(cryptErr.ErrCode), cryptErr.ErrMsg))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request"))
		return
	}
	var msgContent MsgContent
	err = xml.Unmarshal(msg, &msgContent)
	if nil != err {
		s.logger.Error(flag, fmt.Sprintf("decrypt data fail: %s", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request"))
		return
	}
	go s.callback(flag, msgContent, service)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handlerVerify(w http.ResponseWriter, r *http.Request, flag string, service *Service) {
	urlValues := r.URL.Query()
	var (
		verifyMsgSign   = urlValues.Get("msg_signature")
		verifyTimestamp = urlValues.Get("timestamp")
		verifyNonce     = urlValues.Get("nonce")
		verifyEchoStr   = urlValues.Get("echostr")
	)
	wxCrypt := wxbizjsonmsgcrypt.NewWXBizMsgCrypt(service.Token, service.EncodingAesKey, service.AppID, wxbizjsonmsgcrypt.JsonType)
	echoStr, cryptErr := wxCrypt.VerifyURL(verifyMsgSign, verifyTimestamp, verifyNonce, verifyEchoStr)
	if cryptErr != nil {
		s.logger.Error(flag, fmt.Sprintf("verifyURL fail: [%s] %s", strconv.Itoa(cryptErr.ErrCode), cryptErr.ErrMsg))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request"))
	} else {
		s.logger.Info(flag, fmt.Sprintf("verifyURL success: echostr: %s", echoStr))
		w.WriteHeader(http.StatusOK)
		w.Write(echoStr)
	}
}

func (s *Server) callback(flag string, msgContent MsgContent, service *Service) {
	rawContent, err := json.Marshal(msgContent)
	if err != nil {
		s.logger.Error(flag, fmt.Sprintf("marshal msgContent fail: %s", err))
		return
	}
	ctx, cancel := context.WithTimeout(s.context, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, service.Callback, bytes.NewBuffer(rawContent))
	if err != nil {
		s.logger.Error(flag, fmt.Sprintf("create new request fail: %s", err))
		return
	}
	if req.Header == nil {
		req.Header = http.Header{}
	}
	if service.CallbackHeader != nil {
		for k, v := range service.CallbackHeader {
			req.Header.Set(k, v)
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.logger.Error(flag, fmt.Sprintf("do request fail: %s", err))
		return
	}
	if resp.StatusCode != http.StatusOK {
		s.logger.Error(flag, fmt.Sprintf("callback fail: %s", resp.Status))
		return
	}
	s.logger.Info(flag, fmt.Sprintf("callback success: %s", resp.Status))
}
