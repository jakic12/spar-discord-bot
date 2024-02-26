package scraping

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func GetWeeklyMenuImageUrlFromSpar(year int, week int) string {
	/*
		res, err := http.Get(os.Getenv("BASE_MENU_URL"))
		check(err)
		defer res.Body.Close()

		fmt.Printf("[DONE]{%s} %s\n", res.Status, os.Getenv("BASE_MENU_URL"))

		bytes, _ := io.ReadAll(res.Body)
		err = os.WriteFile("data/base.html", bytes, 0644)
		check(err)
		return ""

		doc, err := goquery.NewDocumentFromReader(res.Body)
		check(err)

		url := ""
		doc.Find("img[alt*='jedilnik']").Each(func(i int, s *goquery.Selection) {
			val, err := s.Attr("src")
			if !err {
				if strings.Contains(val, "/jcr:") {
					url = strings.Split(val, "/jcr:")[0]
				} else {
					url = val
				}
			}
		})

		return url
	*/

	return fmt.Sprintf("https://www.spar.si/content/dam/sparsiwebsite/restavracija-interspar/jedilniki/03-24/jedilnik-%02d-%02d.jpg", week, year-(year/1000)*1000)
}

func GetWeeklyMenuImage(imageURL string, imagePath string) {
	res, err := http.Get(imageURL)
	check(err)
	defer res.Body.Close()

	fmt.Printf("[DONE]{%s} %s\n", res.Status, imageURL)

	bytes, _ := io.ReadAll(res.Body)
	err = os.WriteFile(imagePath, bytes, 0644)
	check(err)
}

func GetSubdividedMenus(data_path string, year int, week int) {
	menu_img_path := fmt.Sprintf("%s%02d.jpg", data_path, week)
	if _, err := os.Stat(menu_img_path); errors.Is(err, os.ErrNotExist) {
		weekly_url := GetWeeklyMenuImageUrlFromSpar(year, week)
		GetWeeklyMenuImage(weekly_url, menu_img_path)
	} else if err != nil {
		panic(err)
	}

	sub_images_folder := fmt.Sprintf("%s%02d/", data_path, week)

	if _, err := os.Stat(sub_images_folder); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(sub_images_folder, os.ModePerm); err != nil {
			panic(err)
		}

		image_file, err := os.Open(menu_img_path)
		check(err)

		big_menu, err := jpeg.Decode(image_file)
		for i := 1; i < 7; i++ {
			small_menu, err := getSubMenuForDay(big_menu, i)
			writeImage(small_menu, fmt.Sprintf("%s/%d.jpg", sub_images_folder, i))
			check(err)
		}
		check(err)
	} else if err != nil {
		panic(err)
	}
}

func getSubMenuForDay(img image.Image, day int) (image.Image, error) {
	if day == int(time.Sunday) {
		return img, errors.New("spar isn't open on sunday")
	}

	img_bounds := img.Bounds()
	out_x := img_bounds.Dx() / 3
	out_y := img_bounds.Dy() / 2

	day_coef := day - 1
	out_x_start := out_x * (day_coef % 3)
	out_y_start := out_y * (day_coef / 3)

	return img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(out_x_start, out_y_start, out_x_start+out_x, out_y_start+out_y)), nil
}

func GetImagePathFromDate(data string, date time.Time) string {
	_, week := date.Local().ISOWeek()
	weekday := date.Local().Weekday()

	return fmt.Sprintf("%s%02d/%d.jpg", data, week, weekday)
}

func writeImage(img image.Image, name string) error {
	fd, err := os.Create(name)
	if err != nil {
		return err
	}
	defer fd.Close()

	return jpeg.Encode(fd, img, nil)
}
