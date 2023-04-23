package repository

type WidgetPrototype struct {
	Type          string                 `json:"type"`
	ContainerType string                 `json:"containerType"`
	DisplayName   string                 `json:"displayName"`
	ParentNode    string                 `json:"parentNode"`
	ChildrenNode  []*WidgetPrototype     `json:"childrenNode"`
	H             float64                `json:"h"`
	W             float64                `json:"w"`
	X             float64                `json:"x"`
	Y             float64                `json:"y"`
	Props         map[string]interface{} `json:"props"`
}

func NewWidgetPrototypeByMap(rawWidget map[string]interface{}) *WidgetPrototype {
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
	}
	return widget
}

func (p *WidgetPrototype) AppendChildrenNode(node *WidgetPrototype) {
	p.ChildrenNode = append(p.ChildrenNode, node)
}
