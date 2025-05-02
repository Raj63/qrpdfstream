package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Raj63/qrpdfstream/layout"
	"github.com/Raj63/qrpdfstream/pdf"
)

func main() {
	outFile, err := os.Create("./qrstreamed_1k.pdf")
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	layoutResult := layout.CalculateLayout(layout.PageLayoutParams{
		QRCodeSize: 80,
		HeaderSize: 50,
		FooterSize: 50,
		MarginX:    50,
		MarginY:    50,
		Spacing:    20,
		PageWidth:  595,
		PageHeight: 842,
	})
	p := pdf.NewPDF(outFile, layoutResult)
	var pageIDs []int
	// A4 page dimensions in points

	// Use the calculated values
	fmt.Printf("QR Code Size: %d\n", layoutResult.QRCodeSize)
	fmt.Printf("Columns: %d\n", layoutResult.ColumnCount)
	fmt.Printf("Images per page: %d\n", layoutResult.ImagesPerPage)

	measureExecution(func() {
		logoImg, err := pdf.LoadImageFromFile("./../../assets/qrcite.png")
		if err != nil {
			log.Fatal(err)
		}

		// Add header
		headerStr, headerImgID := p.GenerateHeaderFooterContent("QRCite Report", "Smart Labels", logoImg, true)
		footerStr, footerImgID := p.GenerateHeaderFooterContent("", "Thanks!", nil, false)

		qrcodeSize := layoutResult.QRCodeSize       // 80    // qrcode size
		columnCount := layoutResult.ColumnCount     // 5    // number of columns possible for above qrcodeSize
		imagesPerPage := layoutResult.ImagesPerPage // 35               // Number of QR codes per page possible for above qrcodeSize & columnCount
		// now can we make the qrcodeSize dynamic and then calculate the columnCount and imagesPerPage based on it

		totalQRCodes := 100

		dataList := make([]string, totalQRCodes)
		for i := 0; i < totalQRCodes; i++ {
			dataList[i] = fmt.Sprintf("https://qrcite.com?id=%d", i+1)
		}

		imageIDs := pdf.GenerateQRCodesParallel(p, dataList, qrcodeSize)

		// Stream page generation immediately
		currentPageImages := []int{}

		for i, imgID := range imageIDs {
			currentPageImages = append(currentPageImages, imgID)
			if len(currentPageImages) >= imagesPerPage || i == len(imageIDs)-1 {
				pageID := p.AddPageWithImages(currentPageImages, qrcodeSize, columnCount, headerStr, footerStr, []int{headerImgID, footerImgID})
				pageIDs = append(pageIDs, pageID)
				currentPageImages = []int{} // Reset for the next page
			}
		}

		p.Generate(pageIDs)

		fmt.Println("PDF generated ✅")
	})
}

func measureExecution(fn func()) {
	start := time.Now()
	fn()
	duration := time.Since(start)
	fmt.Printf("Execution Time: %v\n", duration)
}
