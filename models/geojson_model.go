package models

type GeoJSONGeometry struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

type GeoJSONFeature struct {
	Type     string          `json:"type"`
	Geometry GeoJSONGeometry `json:"geometry"`
}

type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}
