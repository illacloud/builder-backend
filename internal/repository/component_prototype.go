package repository

import "math"

const (
	SIZE_REDUCER_TRIGGER_W_LIMIT = 60
)

type WidgetPrototype struct {
	Type           string                 `json:"type"`
	ContainerType  string                 `json:"containerType"`
	DisplayName    string                 `json:"displayName"`
	ParentNode     string                 `json:"parentNode"`
	ChildrenNode   []*WidgetPrototype     `json:"childrenNode"`
	H              float64                `json:"h"`
	W              float64                `json:"w"`
	X              float64                `json:"x"`
	Y              float64                `json:"y"`
	Z              float64                `json:"z"`
	Props          map[string]interface{} `json:"props"`
	IsDragging     bool                   `json:"isDragging"`
	IsResizing     bool                   `json:"isResizing"`
	VerticalResize bool                   `json:"verticalResize"`
	MinH           float64                `json:"minH"`
	MinW           float64                `json:"minW"`
	UnitH          float64                `json:"unitH"`
	UnitW          float64                `json:"unitW"`
}

func NewWidgetPrototypeByMap(rawWidget map[string]interface{}) *WidgetPrototype {
	// assign data
	typeAsserted, _ := rawWidget["type"].(string)
	containerTypeAsserted, _ := rawWidget["containerType"].(string)
	displayNameAsserted, _ := rawWidget["displayName"].(string)
	parentNodeAsserted, _ := rawWidget["parentNode"].(string)
	hAsserted, _ := rawWidget["h"].(float64)
	wAsserted, _ := rawWidget["w"].(float64)
	xAsserted, _ := rawWidget["x"].(float64)
	yAsserted, _ := rawWidget["y"].(float64)
	propsAsserted, _ := rawWidget["props"].(map[string]interface{})
	widget := &WidgetPrototype{
		Type:          typeAsserted,
		ContainerType: containerTypeAsserted,
		DisplayName:   displayNameAsserted,
		ParentNode:    parentNodeAsserted,
		ChildrenNode:  make([]*WidgetPrototype, 0),
		H:             hAsserted,
		W:             wAsserted,
		X:             xAsserted,
		Y:             yAsserted,
		Props:         propsAsserted,
		// reserved field
		IsDragging:     false,
		IsResizing:     false,
		VerticalResize: false,
		MinH:           2,
		MinW:           2,
		UnitH:          hAsserted * 2,
		UnitW:          wAsserted * 2,
		Z:              0,
	}
	return widget
}

func (p *WidgetPrototype) AppendChildrenNode(node *WidgetPrototype) {
	p.ChildrenNode = append(p.ChildrenNode, node)
}

func (p *WidgetPrototype) CheckIfNeedReduceSize() {
	if p.W > SIZE_REDUCER_TRIGGER_W_LIMIT {
		p.H = math.Round(p.H / 5)
		p.W = math.Round(p.W / 5)
		p.X = math.Round(p.X / 5)
		p.Y = math.Round(p.Y / 5)
	}
}
