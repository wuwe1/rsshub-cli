package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	flags "github.com/jessevdk/go-flags"
	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
)

var Reset = "\033[0m"
var Blue = "\033[34m"

const (
	appVersion = "0.0.1"
	appName    = "rsshub"
)

func init() {
	if runtime.GOOS == "windows" {
		Reset = ""
		Blue = ""
	}
}

type options struct {
	Version bool `short:"V" long:"version" description:"Show version"`
}

var opts options

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = appName
	parser.Usage = "[OPTIONS]"
	args, err := parser.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Argument parsing failed.")
		os.Exit(1)
	}

	if opts.Version {
		fmt.Printf("%s: v%s\n", appName, appVersion)
	}

	if len(args) > 0 {
		fmt.Fprintln(os.Stderr, "No arguments required.")
		os.Exit(1)
	}

	routes := fetchRoutes()
	var names []string
	for name := range routes.Data {
		names = append(names, name)
	}

	nameChoices, err := fuzzyfinder.FindMulti(
		names,
		func(i int) string { return names[i] },
		fuzzyfinder.WithPreviewWindow(
			func(i, width, height int) string {
				if i == -1 {
					return ""
				}
				return strings.Join(routes.Data[names[i]].Routes, "\n")
			},
		),
	)

	if err != nil {
		log.Fatal(err)
	}

	var selectedRoutes []string
	for _, nameChoice := range nameChoices {
		selectedRoutes = append(selectedRoutes, routes.Data[names[nameChoice]].Routes...)
	}

	routeChoices, err := fuzzyfinder.FindMulti(
		selectedRoutes,
		func(i int) string { return selectedRoutes[i] },
	)

	if err != nil {
		log.Fatal(err)
	}

	for _, routeChoice := range routeChoices {
		routePath := completeRoute(selectedRoutes[routeChoice])

		feedURL := "https://rsshub.app" + routePath
		feed := fetchFeed(feedURL)

		title := feed.Title
		link := feed.Link
		items := feed.Items

		choices, err := fuzzyfinder.FindMulti(
			items,
			func(i int) string { return items[i].Title },
			fuzzyfinder.WithPreviewWindow(func(i, width, height int) string {
				if i == -1 {
					return ""
				}
				return fmt.Sprintf(
					"%s\n%s\n\nitem: %s\n\nlink: %s\n\ndesc: %s",
					title,
					link,
					items[i].Title,
					items[i].Link,
					items[i].Description,
				)
			}),
		)
		if err != nil {
			log.Fatal(err)
		}

		for _, idx := range choices {
			item := items[idx]
			fmt.Printf("%s\n    %s\n", item.Title, item.Link)
		}
	}
}

func completeRoute(route string) string {
	if strings.Contains(route, ":") {
		tokens := strings.Split(route, "/")
		fmt.Println(route)
		fmt.Println("Please check " +
			Blue + "https://docs.rsshub.app/" + Reset +
			" and fill the parameters.")
		params := make(map[string]string)
		for _, token := range tokens {
			if strings.HasPrefix(token, ":") {
				fmt.Print(strings.TrimPrefix(token, ":") + ": ")
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				input := scanner.Text()
				input = strings.Replace(input, "\n", "", -1)
				input = strings.TrimSpace(input)

				if input == "" && !strings.HasSuffix(token, "?") {
					log.Fatal(fmt.Sprintf("%s is required.\n", strings.TrimPrefix(token, ":")))
				}
				params[token] = input
			}
		}

		for old, new := range params {
			route = strings.Replace(route, old, new, 1)
		}
	}

	return route
}
