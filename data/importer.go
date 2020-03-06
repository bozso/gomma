package data

type ParamKeys struct {
    RngKey, AziKey, TypeKey, DateKey string
}

var (
    DefaultKeys = ParamKeys{
        RngKey: "range_samples",
        AziKey: "azimuth_lines",
        TypeKey: "image_format",
        DateKey: "date",
    }
)
