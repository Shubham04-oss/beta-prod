package ucp

// UCPFeed represents the root array of products in a Universal Commerce Protocol feed.
// It is designed to be easily ingested by AI shopping agents (like Google Gemini).
type UCPFeed []UCPProduct

// UCPProduct represents a single product or product group with its variants in the UCP schema (based on schema.org/Product).
type UCPProduct struct {
	Context     string     `json:"@context"` // Should be "https://schema.org/"
	Type        string     `json:"@type"`    // Should be "Product" or "ProductGroup"
	ProductID   string     `json:"productID"` // Internal ID
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Image       []string   `json:"image,omitempty"`
	Brand       *UCPBrand  `json:"brand,omitempty"`
	Category    string     `json:"category,omitempty"`
	
	// HasVariant holds the specific sellable variants (e.g. Size/Color permutations)
	HasVariant []UCPProductVariant `json:"hasVariant,omitempty"`

	// Offers for single-variant products
	Offers *UCPOffer `json:"offers,omitempty"`
}

type UCPBrand struct {
	Type string `json:"@type"` // "Brand"
	Name string `json:"name"`
}

type UCPProductVariant struct {
	Type        string     `json:"@type"` // "Product"
	SKU         string     `json:"sku"`
	GTIN        string     `json:"gtin,omitempty"`
	Name        string     `json:"name"`
	Image       []string   `json:"image,omitempty"`
	Description string     `json:"description,omitempty"`
	Offers      *UCPOffer  `json:"offers,omitempty"`

	// AdditionalProperty maps to EAV attributes
	AdditionalProperty []UCPPropertyValue `json:"additionalProperty,omitempty"`
}

type UCPPropertyValue struct {
	Type  string `json:"@type"` // "PropertyValue"
	Name  string `json:"name"`  // e.g., "Color"
	Value string `json:"value"` // e.g., "Red"
}

type UCPOffer struct {
	Type          string `json:"@type"` // "Offer"
	Price         string `json:"price"`
	PriceCurrency string `json:"priceCurrency"`
	Availability  string `json:"availability"` // e.g., "https://schema.org/InStock"
	URL           string `json:"url,omitempty"`
}
