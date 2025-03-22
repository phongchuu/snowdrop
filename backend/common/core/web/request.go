package web

import (
	"net/http"
	"strconv"
)

// GetPage retrieves the page number from the URL query parameters of the given HTTP request.
// If the "page" parameter is not provided or cannot be converted to an integer,
// it returns the default page number (DefaultPage) which is 1.
// If the retrieved page number is less than the default page number,
// it also returns the default page number.
func GetPage(r *http.Request) int {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))

	if err != nil {
		return defaultPage
	}

	if page < defaultPage {
		return defaultPage
	}

	return page
}

// GetPageSize retrieves the page size from the URL query parameters of the given HTTP request.
// If the "pageSize" parameter is not provided or cannot be converted to an integer,
// it returns the default page size which is 10.
// If the retrieved page size is less than 1, it returns the default page size.
// If the retrieved page size is greater than the maximum page size (100),
// it returns the maximum page size.
func GetPageSize(r *http.Request) int {
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))

	if err != nil {
		return defaultPageSize
	}

	if pageSize < 1 {
		return defaultPageSize
	}

	if pageSize > maxPageSize {
		return maxPageSize
	}

	return pageSize
}
