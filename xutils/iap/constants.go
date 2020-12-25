package iap

import "errors"

// 验证地址
const (
	sandboxVerifyUrl = `https://sandbox.itunes.apple.com/verifyReceipt`
	backendVerifyUrl = `https://buy.itunes.apple.com/verifyReceipt`
)

// 响应码
const (
	// 验证成功
	respCodeSuccess = 0

	// AppStore 无法读取您提供的JSON对象
	respCode21000 = 21000

	// receipt-data 属性中的数据格式错误或丢失
	respCode21002 = 21002

	// 无法认证收据
	respCode21003 = 21003

	// 您提供的共享密钥与您账户存档的共享密钥不匹配
	respCode21004 = 21004

	// 收据服务器当前不可用
	respCode21005 = 21005

	// 此收据有效，但订阅已过期。当此状态码返回到您的服务器时，收据数据也将解码并作为响应的一部分返回。
	// 只有在交易收据为iOS 6样式且为自动续期订阅时才会返回
	respCode21006 = 21006

	// 此收据来自测试环境，但发送到生产环境进行验证。应将其发送到测试环境。
	respCode21007 = 21007

	// 此收据来自生产环境，但发送到测试环境进行验证。应将其发送到生产环境。
	respCode21008 = 21008

	// 此收据无法获得授权。对待此收据的方式与从未进行过任何交易时的处理方式相同。
	respCode21010 = 21010
)

// 验证错误定义
var (
	err21000  = errors.New(`AppStore 无法读取您提供的JSON对象`)
	err21002  = errors.New(`receipt-data 属性中的数据格式错误或丢失`)
	err21003  = errors.New(`无法认证收据`)
	err21004  = errors.New(`您提供的共享密钥与您账户存档的共享密钥不匹配`)
	err21005  = errors.New(`收据服务器当前不可用`)
	err21006  = errors.New(`此收据有效，但订阅已过期`)
	err21007  = errors.New(`此收据来自测试环境，但发送到生产环境进行验证。应将其发送到测试环境`)
	err21008  = errors.New(`此收据来自生产环境，但发送到测试环境进行验证。应将其发送到生产环境`)
	err21010  = errors.New(`此收据无法获得授权。对待此收据的方式与从未进行过任何交易时的处理方式相同。`)
	errOthers = errors.New(`内部数据访问错误`)
)
