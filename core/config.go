package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"wxcallback/lib/log"
	"wxcallback/lib/types"
)

type Config struct {
	Mode    string                  `json:"mode,omitempty"`
	Listen  string                  `json:"listen,omitempty"`
	Service types.Listable[Service] `json:"service,omitempty"`
}

type Service struct {
	Listen         string            `json:"listen,omitempty"`
	Path           string            `json:"path,omitempty"`
	Token          string            `json:"token,omitempty"`
	AppID          string            `json:"app_id,omitempty"`
	AgentID        int               `json:"agent_id,omitempty"`
	Secret         string            `json:"secret,omitempty"`
	EncodingAesKey string            `json:"encoding_aes_key,omitempty"`
	VerifyUrl      bool              `json:"verify_url,omitempty"`
	Callback       string            `json:"callback,omitempty"`
	CallbackHeader map[string]string `json:"callback_header,omitempty"`
}

type _config Config

func (c *Config) UnmarshalJSON(content []byte) error {
	var _c _config
	err := json.Unmarshal(content, &_c)
	if err != nil {
		return err
	}
	err = checkConfig(&_c)
	if err != nil {
		return err
	}
	*c = Config(_c)
	return nil
}

func checkConfig(_c *_config) error {
	switch _c.Mode {
	case "port":
		if len(_c.Service) == 0 {
			return fmt.Errorf("service is empty")
		}
		checkMap := make(map[string]int)
		services := make(types.Listable[Service], 0)
		for _, service := range _c.Service {
			if service.Listen == "" {
				return fmt.Errorf("listen is empty")
			} else {
				if _, ok := checkMap[fmt.Sprintf("listen:%s", service.Listen)]; ok {
					return fmt.Errorf("duplicate listen: %s", service.Listen)
				} else {
					checkMap[fmt.Sprintf("listen:%s", service.Listen)]++
				}
			}
			if service.Token == "" {
				return fmt.Errorf("listen: %s, token is empty", service.Listen)
			}
			if service.AppID == "" {
				return fmt.Errorf("listen: %s, appid is empty", service.Listen)
			}
			if service.AgentID <= 0 {
				return fmt.Errorf("listen: %s, agentid is empty", service.Listen)
			}
			if service.Secret == "" {
				return fmt.Errorf("listen: %s, secret is empty", service.Listen)
			}
			if service.EncodingAesKey == "" {
				return fmt.Errorf("listen: %s, encoding_aes_key is empty", service.Listen)
			}
			if service.Callback == "" {
				return fmt.Errorf("listen: %s, callback is empty", service.Listen)
			} else {
				u, err := url.Parse(service.Callback)
				if err != nil {
					_, _, err = net.SplitHostPort(service.Callback)
					if err != nil {
						return fmt.Errorf("listen: %s, invaild format callback url: %s", service.Listen, err)
					}

				}
				if u.Scheme != "http" && u.Scheme != "https" {
					return fmt.Errorf("listen: %s, callback only support http/https", service.Listen)
				}
			}
			if len(service.CallbackHeader) == 0 {
				service.CallbackHeader = nil
			}
			services = append(services, service)
		}
		_c.Service = services
	case "", "path":
		_c.Mode = "path"
		if _c.Listen == "" {
			return fmt.Errorf("listen is empty")
		}
		if len(_c.Service) == 0 {
			return fmt.Errorf("service is empty")
		}
		checkMap := make(map[string]int)
		services := make(types.Listable[Service], 0)
		for _, service := range _c.Service {
			if service.Path != "" {
				if service.Path[0] != '/' {
					service.Path = "/" + service.Path
				}
			} else {
				service.Path = "/"
			}
			if _, ok := checkMap[fmt.Sprintf("path:%s", service.Path)]; ok {
				return fmt.Errorf("duplicate path: %s", service.Path)
			} else {
				checkMap[fmt.Sprintf("path:%s", service.Path)]++
			}
			if service.Token == "" {
				return fmt.Errorf("path: %s, token is empty", service.Path)
			}
			if service.AppID == "" {
				return fmt.Errorf("path: %s, appid is empty", service.Path)
			}
			if service.AgentID <= 0 {
				return fmt.Errorf("path: %s, agentid is empty", service.Path)
			}
			if service.Secret == "" {
				return fmt.Errorf("path: %s, secret is empty", service.Path)
			}
			if service.EncodingAesKey == "" {
				return fmt.Errorf("path: %s, encoding_aes_key is empty", service.Path)
			}
			if service.Callback == "" {
				return fmt.Errorf("path: %s, callback is empty", service.Path)
			} else {
				u, err := url.Parse(service.Callback)
				if err != nil {
					_, _, err = net.SplitHostPort(service.Callback)
					if err != nil {
						return fmt.Errorf("path: %s, invaild format callback url: %s", service.Path, err)
					}

				}
				if u.Scheme != "http" && u.Scheme != "https" {
					return fmt.Errorf("path: %s, callback only support http/https", service.Path)
				}
			}
			if len(service.CallbackHeader) == 0 {
				service.CallbackHeader = nil
			}
			services = append(services, service)
		}
		_c.Service = services
	default:
		return fmt.Errorf("invalid mode: %s", _c.Mode)
	}
	return nil
}

type Server struct {
	config Config
	//
	context context.Context
	logger  *log.Logger
}

type ServerOption struct {
	Context context.Context
	Logger  *log.Logger
}
