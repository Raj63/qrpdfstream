package layout

type PageLayout struct {
	QRCodeSize    int
	ColumnCount   int
	ImagesPerPage int
}

// A4 page dimensions in points
const (
	pageWidth  = 595
	pageHeight = 842
	marginX    = 50
	marginY    = 50
	spacing    = 20
)

// Calculates column count and images per page based on QR code size
func CalculateLayout(qrcodeSize int) PageLayout {
	usableWidth := pageWidth - 2*marginX
	usableHeight := pageHeight - 2*marginY

	columnCount := usableWidth/(qrcodeSize+spacing) + 1
	rows := usableHeight / (qrcodeSize + spacing)

	imagesPerPage := columnCount * rows

	return PageLayout{
		QRCodeSize:    qrcodeSize,
		ColumnCount:   columnCount,
		ImagesPerPage: imagesPerPage,
	}
}
