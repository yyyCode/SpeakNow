package xunfei

import "strings"

// Decoder 解析讯飞流式返回的分片结果（支持 rpl 替换）。
type Decoder struct {
	results []*wsResult
}

func (d *Decoder) Decode(result *wsResult) {
	if result == nil {
		return
	}
	if len(d.results) <= result.Sn {
		need := result.Sn - len(d.results) + 1
		d.results = append(d.results, make([]*wsResult, need)...)
	}
	if result.Pgs == "rpl" && len(result.Rg) == 2 {
		for i := result.Rg[0]; i <= result.Rg[1]; i++ {
			if i < len(d.results) {
				d.results[i] = nil
			}
		}
	}
	d.results[result.Sn] = result
}

func (d *Decoder) String() string {
	var b strings.Builder
	for _, r := range d.results {
		if r == nil {
			continue
		}
		b.WriteString(r.text())
	}
	return b.String()
}

type wsResult struct {
	Ls  bool   `json:"ls"`
	Rg  []int  `json:"rg"`
	Sn  int    `json:"sn"`
	Pgs string `json:"pgs"`
	Ws  []wsWord `json:"ws"`
}

func (r *wsResult) text() string {
	var b strings.Builder
	for _, w := range r.Ws {
		for _, cw := range w.Cw {
			b.WriteString(cw.W)
		}
	}
	return b.String()
}

type wsWord struct {
	Bg int    `json:"bg"`
	Cw []wsCw `json:"cw"`
}

type wsCw struct {
	Sc int    `json:"sc"`
	W  string `json:"w"`
}

type wsResp struct {
	Sid     string `json:"sid"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Result wsResult `json:"result"`
		Status int      `json:"status"`
	} `json:"data"`
}
