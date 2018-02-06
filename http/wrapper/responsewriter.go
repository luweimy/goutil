package wrapper

import (
	"net/http"
	"net/http/httptest"
	"sync"
)

type ResponseWriterModifyHeaderWrapper struct {
	sync.Once
	httptest.ResponseRecorder
	responseWriter http.ResponseWriter
}

// 解决HttpResponseHeader修改不生效的问题
//
// http.ResponseWriter.WriteHeader(写入状态码)调用后再修改ResponseHeader就不再生效，
// 原因是WriteHeader时内部给Header做了一个快照，所以其后修改的Header能打印出来，但是不会写回到客户端。
//
// 调用http.response.Write时内部又会自动调用response.WriteHeader，所以导致包装Wrapper拦截WriteHeader延迟写入时，
// 只能拦截通过ResponseWriter对WriteHeader的调用，但无法拦截response内部对WriteHeader的调用。
//
// 此处用httptest.ResponseRecorder作为代替，让所有的Response数据临时写入ResponseRecorder，
// 然后拷贝回到http.ResponseWriter中，此时Header拷贝可以先于WriteHeader，对Header的修改就可以全部生效。
func NewResponseWriterModifyHeaderWrapper(w http.ResponseWriter) *ResponseWriterModifyHeaderWrapper {
	return &ResponseWriterModifyHeaderWrapper{
		ResponseRecorder: *httptest.NewRecorder(),
		responseWriter:   w,
	}
}

func (wrapper *ResponseWriterModifyHeaderWrapper) Done() error {
	var err error
	wrapper.Do(func() {
		// 从Recorder向Response拷贝Header
		copyHeaders(wrapper.responseWriter.Header(), wrapper.Header())

		// WriteHeader调用之后ResponseHeader不可修改
		// 要在responseWriter.Write之前，否则responseWriter.Write内部会自动设置状态码200
		wrapper.responseWriter.WriteHeader(wrapper.Code)

		// 从Recorder向Response拷贝Body
		data := wrapper.Body.Bytes()
		_, err = wrapper.responseWriter.Write(data)
	})
	return err
}

func copyHeaders(dst, src http.Header) {
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}
