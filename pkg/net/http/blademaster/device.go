package blademaster

import "net/http"

const (
	// 公用的HEADER
	_httpHeaderUA             = "User-Agent"      // boohee/ios
	_httpHeaderOsVersion      = "Os-Version"      // 手机系统版本号
	_httpHeaderUserAgent      = "UserAgent"       // 用户代理
	_httpHeaderAppDevice      = "App-Device"      // 操作系统
	_httpHeaderPhoneModel     = "Phone-Model"     // 设备型号
	_httpHeaderUTCOffset      = "Utc-Offset"      // 与GMT时间偏移秒数
	_httpHeaderDeviceToken    = "Device-Token"    // 设备标识
	_httpHeaderAnonymousToken = "Anonymous-Token" // 游客态token
	_httpHeaderAppVersion     = "App-Version"     // 版本号
	_httpHeaderAppKey         = "App-Key"         // 薄荷健康为 one  薄荷营养师 food

	// iOS/Android 差异化
	_httpHeaderAcceptLanguage = "Accept-Language"
	_httpHeaderPhonePlatform  = "Phone-Platform"
	_httpHeaderPhoneDevice    = "Phone-Device" // 系统返回设备型号
	_httpHeaderVersionCode    = "Version-Code" // 与App-Version是配对的，一般用于应用市场更新
	_httpHeaderChannel        = "channel"      // 下载渠道
	_httpHeaderOAID           = "OAID"         // 设备标识
	_httpHeaderUserKey        = "User-Key"
)

type Device struct {
	UserAgent      string `json:"User-Agent"`
	OsVersion      string `json:"Os-Version"`
	UA             string `json:"UserAgent"`
	AppDevice      string `json:"App-Device"`
	PhoneModel     string `json:"Phone-Model"`
	UTCOffset      string `json:"Utc-Offset"`
	DeviceToken    string `json:"Device-Token"`
	AnonymousToken string `json:"Anonymous-Token"`
	AppVersion     string `json:"App-Version"`
	AppKey         string `json:"App-Key"`
	AcceptLanguage string `json:"Accept-Language"`
	PhonePlatform  string `json:"Phone-Platform"`
	PhoneDevice    string `json:"Phone-Device"`
	VersionCode    string `json:"Version-Code"`
	Channel        string `json:"channel"`
	OAID           string `json:"OAID"`
	UserKey        string `json:"User-Key"`
}

func DeviceInfo(req *http.Request) (device *Device) {
	device = &Device{}
	device.UserAgent = req.Header.Get(_httpHeaderUA)
	device.OsVersion = req.Header.Get(_httpHeaderOsVersion)
	device.UA = req.Header.Get(_httpHeaderUserAgent)
	device.AppDevice = req.Header.Get(_httpHeaderAppDevice)
	device.PhoneModel = req.Header.Get(_httpHeaderPhoneModel)
	device.UTCOffset = req.Header.Get(_httpHeaderUTCOffset)
	device.DeviceToken = req.Header.Get(_httpHeaderDeviceToken)
	device.AnonymousToken = req.Header.Get(_httpHeaderAnonymousToken)
	device.AppVersion = req.Header.Get(_httpHeaderAppVersion)
	device.AppKey = req.Header.Get(_httpHeaderAppKey)
	device.AcceptLanguage = req.Header.Get(_httpHeaderAcceptLanguage)
	device.PhonePlatform = req.Header.Get(_httpHeaderPhonePlatform)
	device.PhoneDevice = req.Header.Get(_httpHeaderPhoneDevice)
	device.VersionCode = req.Header.Get(_httpHeaderVersionCode)
	device.Channel = req.Header.Get(_httpHeaderChannel)
	device.OAID = req.Header.Get(_httpHeaderOAID)
	device.UserKey = req.Header.Get(_httpHeaderUserKey)
	return
}
