package config

import "github.com/Unknwon/goconfig"

var GlobalConf *Config
var cfg *goconfig.ConfigFile

type Config struct {
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"database"`

	Server struct {
		Port           int    `json:"port"`
		Debug          bool   `json:"debug"`
		Proxy          bool   `json:"proxy"`
		ProxyUrl       string `json:"proxyUrl"`
		InitExpireTime int    `json:"initExpireTime"`
	} `json:"server"`

	OpenAi struct {
		AppKeys    []string `json:"appkeys"`
		Url        string   `json:"gpt3_5TurboUrl"`
		ChatLimit  int      `json:"chatLimit"`
		VipAppKeys []string `json:"vipAppkeys"`
	}

	SMS struct {
		TemplateCode  string   `json:"templateCode"`
		Ak            string   `json:"ak"`
		Sk            string   `json:"sk"`
		Sign          string   `json:"sign"`
		EmailFrom     string   `json:"emailFrom"`
		BlackoutEmail []string `json:"blackoutEmail"`
		Frequent      int      `json:"frequent"`
	}

	Pay struct {
		AppId        string `json:"appid"`
		PrivateKey   string `json:"privateKey"`
		IsProduction bool   `json:"isProduction"`
		AliPublicKey string `json:"aliPublicKey"`
		Timeout      string `json:"timeout"`
		NotifyUrl    string `json:"notifyURL"`
		ReturnUrl    string `json:"returnURL"`
	} `json:"pay"`

	About struct {
		DownloadUrl             string `json:"downloadUrl"`
		ChromePluginDownloadUrl string `json:"pluginDownloadUrl"`
		Version                 string `json:"version"`
		ChromePluginVersion     string `json:"pluginVersion"`
		Password                string `json:"password"`
		MailTo                  string `json:"mailTo"`
		Info                    string `json:"info"`
		DocumentUrl             string `json:"documentUrl"`
		PluginInfo              string `json:"pluginInfo"`
	}

	PDF struct {
		Server     string   `json:"server"`
		Dir        string   `json:"dir"`
		PdfCnt     int      `json:"pdfCnt"`
		QACnt      int      `json:"qaCnt"`
		AppKeys    []string `json:"appkeys"`
		VipAppKeys []string `json:"vipAppkeys"`
		Size       int      `json:"size"`
		Formats    string   `json:"formats"`
		About      string   `json:"about"`
	}

	Chat struct {
		UserQALimitCnt    int `json:"userQALimitCnt"`
		UserImageLimitCnt int `json:"userImageLimitCnt"`
	}
}

func LoadCfg(configFile string) error {
	var err error
	cfg, err = goconfig.LoadConfigFile(configFile)
	return err
}

func LoadConfig() (*Config, error) {
	cfg.Reload()
	var config Config
	// 读取配置文件中的值并赋值给 config 变量
	config.Database.Host = cfg.MustValue("database", "host")
	config.Database.Port = cfg.MustInt("database", "port")
	config.Database.Username = cfg.MustValue("database", "username")
	config.Database.Password = cfg.MustValue("database", "password")
	config.Database.Database = cfg.MustValue("database", "database")

	config.Server.Port = cfg.MustInt("server", "port")
	config.Server.Debug = cfg.MustBool("server", "debug")
	config.Server.Proxy = cfg.MustBool("server", "proxy")
	config.Server.ProxyUrl = cfg.MustValue("server", "proxyUrl")
	config.Server.InitExpireTime = cfg.MustInt("server", "initExpireTime")

	config.OpenAi.AppKeys = cfg.MustValueArray("openai", "appkeys", ",")
	config.OpenAi.Url = cfg.MustValue("openai", "gpt3_5TurboUrl")
	config.OpenAi.ChatLimit = cfg.MustInt("openai", "chatLimit")
	config.OpenAi.VipAppKeys = cfg.MustValueArray("openai", "vipAppkeys", ",")
	config.Pay.AppId = cfg.MustValue("pay", "appid")
	config.Pay.PrivateKey = cfg.MustValue("pay", "privateKey")
	config.Pay.AliPublicKey = cfg.MustValue("pay", "aliPublicKey")
	config.Pay.NotifyUrl = cfg.MustValue("pay", "notifyURL")
	config.Pay.ReturnUrl = cfg.MustValue("pay", "returnURL")
	config.Pay.IsProduction = cfg.MustBool("pay", "isProduction")
	config.Pay.Timeout = cfg.MustValue("pay", "timeout")

	config.SMS.TemplateCode = cfg.MustValue("sms", "templateCode")
	config.SMS.Ak = cfg.MustValue("sms", "ak")
	config.SMS.Sk = cfg.MustValue("sms", "sk")
	config.SMS.Sign = cfg.MustValue("sms", "sign")
	config.SMS.EmailFrom = cfg.MustValue("sms", "emailFrom")
	config.SMS.BlackoutEmail = cfg.MustValueArray("sms", "blackoutEmail", ",")
	config.SMS.Frequent = cfg.MustInt("sms", "frequent", 60)

	config.About.DownloadUrl = cfg.MustValue("about", "downloadUrl")
	config.About.ChromePluginDownloadUrl = cfg.MustValue("about", "chromePluginDownloadUrl")
	config.About.Version = cfg.MustValue("about", "version")
	config.About.ChromePluginVersion = cfg.MustValue("about", "chromePluginVersion")
	config.About.Password = cfg.MustValue("about", "password")
	config.About.MailTo = cfg.MustValue("about", "mailTo")
	config.About.Info = cfg.MustValue("about", "info")
	config.About.DocumentUrl = cfg.MustValue("about", "documentUrl")
	config.About.PluginInfo = cfg.MustValue("about", "pluginInfo", "")

	config.PDF.Server = cfg.MustValue("pdf", "server")
	config.PDF.Dir = cfg.MustValue("pdf", "dir")
	config.PDF.Formats = cfg.MustValue("pdf", "formats")
	config.PDF.About = cfg.MustValue("pdf", "about")
	config.PDF.PdfCnt = cfg.MustInt("pdf", "pdfCnt")
	config.PDF.QACnt = cfg.MustInt("pdf", "qaCnt")
	config.PDF.AppKeys = cfg.MustValueArray("openai", "appkeys", ",")
	config.PDF.VipAppKeys = cfg.MustValueArray("openai", "vipAppkeys", ",")
	config.PDF.Size = cfg.MustInt("pdf", "size")

	config.Chat.UserQALimitCnt = cfg.MustInt("chat", "userQALimitCnt", 12)
	config.Chat.UserImageLimitCnt = cfg.MustInt("chat", "userImageLimitCnt", 8)

	GlobalConf = &config
	return &config, nil
}
