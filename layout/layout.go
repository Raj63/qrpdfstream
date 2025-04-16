package layout

type PageLayout struct {
	QRCodeSize    int
	ColumnCount   int
	ImagesPerPage int
	HeaderSize    int
	FooterSize    int
	Spacing       int
	PageWidth     int
	PageHeight    int
	MarginX       int
	MarginY       int
}

type PageLayoutParams struct {
	QRCodeSize int
	HeaderSize int
	FooterSize int
	MarginX    int
	MarginY    int
	Spacing    int
	PageWidth  int
	PageHeight int
}

// Calculates column count and images per page based on QR code size
func CalculateLayout(params PageLayoutParams) PageLayout {
	usableWidth := params.PageWidth - 2*params.MarginX
	usableHeight := params.PageHeight - 2*params.MarginY - (params.HeaderSize + params.FooterSize)

	columnCount := usableWidth / (params.QRCodeSize + params.Spacing)
	rows := usableHeight / (params.QRCodeSize + params.Spacing)

	imagesPerPage := columnCount * rows

	return PageLayout{
		QRCodeSize:    params.QRCodeSize,
		ColumnCount:   columnCount,
		ImagesPerPage: imagesPerPage,
		HeaderSize:    params.HeaderSize,
		FooterSize:    params.FooterSize,
		MarginX:       params.MarginX,
		MarginY:       params.MarginY,
		Spacing:       params.Spacing,
		PageWidth:     params.PageWidth,
		PageHeight:    params.PageHeight,
	}
}
