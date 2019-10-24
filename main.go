package main

import (
  "fmt"
  "log"
  "strconv"
  "strings"
  "net/http"
  "io/ioutil"
)

type podcast struct {
  Name string
  Image string
  Episodes []episode
}

type episode struct {
  Filename string
  Size int
}

var podcasts map[string]podcast

var mediaEndings = []string{"mp3", "ogg", "wav", "mp4", "webm"}
var imageEndings = []string{"jpg", "jpeg", "png", "tiff"}

func main() {
  podcasts = make(map[string]podcast)
  scanDirectory("data")

  for _, podcast := range podcasts {
    fmt.Println(podcast)
  }

  http.HandleFunc("/", sendHtml)
  http.HandleFunc("/podcasts/", sendXml)
  http.Handle("/media/" + "data/", http.StripPrefix("/media/" + "data/", http.FileServer(http.Dir("data"))))
  log.Fatal(http.ListenAndServe(":8080", nil))
}

func scanDirectory(directory string) {
  podcast := podcast{Name: directory, Image: "", Episodes: make([]episode, 0)}
  files, _ := ioutil.ReadDir(directory)
  for _, f := range files {
    if(f.IsDir()) {
      scanDirectory(directory + "/" + f.Name())
    } else if checkFileending(f.Name(), mediaEndings) {
      episode := episode{Filename: f.Name(), Size: int(f.Size())}
      podcast.Episodes = append(podcast.Episodes, episode)
    } else if checkFileending(f.Name(), imageEndings) {
      podcast.Image = f.Name()
    }
  }
  podcasts[podcast.Name] = podcast
}

func checkFileending(filename string, endings []string) bool {
  for _, ending := range endings {
    if strings.HasSuffix(strings.ToLower(filename), "." + strings.ToLower(ending)) {
      return true
    }
  }
  return false
}

func sendHtml(w http.ResponseWriter, req *http.Request) {
  for _, podcast := range podcasts {
    fmt.Fprintf(w, "<html><head><title>directory2podcast</title></head><body><ul>\n")
    fmt.Fprintf(w, "<li><a href=\"podcasts/" + podcast.Name + "\">" + podcast.Name + "</a></li>\n")
    fmt.Fprintf(w, "</ul></body></html>\n")
  }
}

func sendXml(w http.ResponseWriter, req *http.Request) {
  podcast, present := podcasts[req.URL.Path[len("/podcasts/"):]]

  if(present) {
    fmt.Fprintf(w, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
    fmt.Fprintf(w, "<rss version=\"2.0\" xmlns:atom=\"http://www.w3.org/2005/Atom\">\n")
    fmt.Fprintf(w, "  <channel>\n")
    fmt.Fprintf(w, "    <title>" + podcast.Name  + "</title>\n")
    fmt.Fprintf(w, "    <image><url>http://" + req.Host + "/media/" + podcast.Name + "/" + podcast.Image + "</url></image>\n")
    fmt.Fprintf(w, "    <generator>directory2podcast</generator>\n")

    for _, episode := range podcast.Episodes {
      fmt.Fprintf(w, "    <item>\n")
      fmt.Fprintf(w, "      <title>" + episode.Filename + "</title>\n")
      fmt.Fprintf(w, "      <link>media/" + episode.Filename + "</link>\n")
      fmt.Fprintf(w, "      <description><![CDATA[" + episode.Filename + "]]></description>\n")
      fmt.Fprintf(w, "      <enclosure url=\"http://" + req.Host + "/media/" + podcast.Name + "/" +  episode.Filename + "\" length=\"" + strconv.Itoa(episode.Size) + "\" type=\"audio/mp3\" />\n")
      fmt.Fprintf(w, "    </item>\n")
    }

    fmt.Fprintf(w, "  </channel>\n")
    fmt.Fprintf(w, "</rss>\n")
  } else {
    sendHtml(w, req)
  }
}
