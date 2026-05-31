package paymentsvc

import stdhttp "net/http"

func cloneHeader(src stdhttp.Header) stdhttp.Header {
	dst := make(stdhttp.Header)
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
	return dst
}

func mergeHeader(dst stdhttp.Header, src stdhttp.Header) {
	for key, values := range src {
		dst.Del(key)
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
