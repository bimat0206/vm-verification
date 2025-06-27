package config

type Config struct {
	BucketName       string // Optional - will be overridden by S3 event
	S3Region         string
	FontPath         string
	BoldFontPath     string
	ImageLoadTimeout int // in seconds
	CanvasScale      float64
	NumColumns       int
	CellWidth        float64
	CellHeight       float64
	RowSpacing       float64
	CellSpacing      float64
	HeaderHeight     float64
	FooterHeight     float64
	ImageSize        float64
	Padding          float64
	TitlePadding     float64
	TextPadding      float64
	MetadataHeight   float64
	TitleFontSize    float64
	HeaderFontSize   float64
	PositionFontSize float64
}

var config = Config{
	BucketName:       "", // Empty by default, will be set from S3 event
	S3Region:         "us-east-1",
	FontPath:         "fonts/arial.ttf",
	BoldFontPath:     "fonts/arialbd.ttf",
	ImageLoadTimeout: 20,
	CanvasScale:      4.0,
	NumColumns:       20,
	CellWidth:        150.0,
	CellHeight:       180.0,
	RowSpacing:       60.0,
	CellSpacing:      10.0,
	HeaderHeight:     40.0,
	FooterHeight:     30.0,
	ImageSize:        100.0,
	Padding:          20.0,
	TitlePadding:     40.0,
	TextPadding:      5.0,
	MetadataHeight:   20.0,
	TitleFontSize:    18.0,
	HeaderFontSize:   14.0,
	PositionFontSize: 14.0,
}

func GetConfig() Config {
	return config
}

func UpdateBucketName(bucketName string) {
	config.BucketName = bucketName
}
