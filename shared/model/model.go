package model

/******************************************************************************
 * 通用结构体
******************************************************************************/
// type ResponseBody map[string]interface{}
type Response struct {
	Msg  string      `json:"msg"`
	Code uint32      `json:"code"`
	Data interface{} `json:"data,omitempty"`
}
