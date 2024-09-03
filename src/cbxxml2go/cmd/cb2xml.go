package main

type Condition struct {
	Name    string      `xml:"name,attr"`
	Through string      `xml:"through,attr"`
	Value   string      `xml:"value,attr"`
	Items   []Condition `xml:"condition"`
}

type Copybook struct {
	Filename string `xml:"filename,attr"`
	Items    []Item `xml:"item"`
}

type Item struct {
	AssumedDigits      int         `xml:"assumed-digits,attr"`
	DependingOn        string      `xml:"depending-on,attr"`
	DisplayLength      int         `xml:"display-length,attr"`
	EdittedNumeric     bool        `xml:"editted-numeric,attr"`
	InheritedUsage     bool        `xml:"inherited-usage,attr"`
	InsertDecimalPoint bool        `xml:"insert-decimal-point,attr"`
	Level              string      `xml:"level,attr"`
	Name               string      `xml:"name,attr"`
	Numeric            string      `xml:"numeric,attr"` //FIXME: this was a bool?? - check this
	Occurs             int         `xml:"occurs,attr"`
	OccursMin          int         `xml:"occurs-min,attr"`
	Picture            string      `xml:"picture,attr"`
	Position           int         `xml:"position,attr"`
	Redefined          string      `xml:"redefined,attr"`
	Redefines          string      `xml:"redefines,attr"`
	Scale              int         `xml:"scale,attr"`
	SignPosition       string      `xml:"sign-position,attr"`
	SignSeparate       bool        `xml:"sign-separate,attr"`
	Signed             bool        `xml:"signed,attr"`
	StorageLength      int         `xml:"storage-length,attr"`
	Sync               bool        `xml:"sync,attr"`
	Usage              string      `xml:"usage,attr"`
	Value              string      `xml:"value,attr"`
	Conditions         []Condition `xml:"condition"`
	Items              []Item      `xml:"item"`
}
