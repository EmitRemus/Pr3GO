package main

import (
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

type PageData struct {
	Sig1 float64
	Sig2 float64
	Ps   float64
	B    float64

	Result map[string]interface{}
}

var pageTmpl = template.Must(template.ParseFiles("index.html"))

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/margin", marginHandler)

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static")),
		),
	)

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/margin", http.StatusSeeOther)
}

func marginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleMarginGet(w, r)
	case http.MethodPost:
		handleMarginPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleMarginGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	data := PageData{
		Sig1: parseFloat(query, "sig1"),
		Sig2: parseFloat(query, "sig2"),
		Ps:   parseFloat(query, "ps"),
		B:    parseFloat(query, "b"),
	}

	if query.Get("sig1") != "" {
		data.Result = calcMargin(data.Sig1, data.Sig2, data.Ps, data.B)
	}

	if err := pageTmpl.Execute(w, data); err != nil {
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
	}
}

func handleMarginPost(w http.ResponseWriter, r *http.Request) {
	values := url.Values{
		"sig1": []string{r.FormValue("sig1")},
		"sig2": []string{r.FormValue("sig2")},
		"ps":   []string{r.FormValue("ps")},
		"b":    []string{r.FormValue("b")},
	}

	http.Redirect(w, r, "/margin?"+values.Encode(), http.StatusSeeOther)
}

func parseFloat(query url.Values, key string) float64 {
	value, _ := strconv.ParseFloat(query.Get(key), 64)
	return value
}

func calcMargin(sig1, sig2, ps, b float64) map[string]interface{} {
	const delta = 0.05
	pMin, pMax := ps-ps*delta, ps+ps*delta

	deltaW1 := math.Round(integrate(pMin, pMax, sig1, ps)*100) / 100
	W1 := int(math.Round(ps * 24 * deltaW1))
	P1 := int(math.Round(float64(W1) * b))
	W2 := int(math.Round(ps * 24 * (1 - deltaW1)))
	H1 := int(math.Round(float64(W2) * b))

	deltaW2 := math.Round(integrate(pMin, pMax, sig2, ps)*100) / 100
	W3 := int(math.Round(ps * 24 * deltaW2))
	P2 := int(math.Round(float64(W3) * b))
	W4 := int(math.Round(ps * 24 * (1 - deltaW2)))
	H2 := int(math.Round(float64(W4) * b))

	return map[string]interface{}{
		"deltaW1": deltaW1 * 100,
		"P1":      P1,
		"H1":      H1,
		"P1-H1":   P1 - H1,
		"deltaW2": deltaW2 * 100,
		"P2":      P2,
		"H2":      H2,
		"P2-H2":   P2 - H2,
	}
}

func integrate(pStart, pEnd, sig, ps float64) float64 {
	toIntegrate := func(sig, ps, p float64) float64 {
		dif := p - ps
		power := math.Pow(dif, 2) / (2 * sig * sig)
		numerator := math.Exp(power)
		denominator := sig * math.Sqrt(6.28)
		return numerator / denominator
	}

	sum := 0.5*toIntegrate(sig, ps, pStart) + 0.5*toIntegrate(sig, ps, pEnd)
	iterations := 1000
	delta := (pEnd - pStart) / float64(iterations)
	for i := 1; i < iterations; i++ {
		sum += toIntegrate(sig, ps, pStart+float64(i)*delta)
	}
	sum *= delta
	return sum
}
