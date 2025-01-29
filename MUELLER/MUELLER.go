package MUELLER

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type Server struct {
	port       int
	routes     map[string]map[string]func() string
	soundMap   map[string]string
	bgMusic    string
	defaultDir string
	mixer      *beep.Mixer
}

func NewServer(port int) *Server {

	return &Server{
		port:       port,
		routes:     make(map[string]map[string]func() string),
		soundMap:   make(map[string]string),
		defaultDir: "server_sounds",
		mixer:      &beep.Mixer{},
	}
}

func (s *Server) Start() {
	fmt.Println("\n   *                        (        (                (     \n (  `                       )\\ )     )\\ )             )\\ )  \n )\\))(        (     (      (()/(    (()/(     (      (()/(  \n((_)()\\       )\\    )\\      /(_))    /(_))    )\\      /(_)) \n(_()((_)   _ ((_)  ((_)    (_))     (_))     ((_)    (_))   \n|  \\/  |  | | | |  | __|   | |      | |      | __|   | _ \\  \n| |\\/| | _| |_| |_ | _|  _ | |__  _ | |__  _ | _|  _ |   /  \n|_|  |_|(_)\\___/(_)|___|(_)|____|(_)|____|(_)|___|(_)|_|_\\\n")
	fmt.Println("The kleine süße Lösung is starting: \n")
	if s.bgMusic == "" {
		s.bgMusic = filepath.Join(s.defaultDir, "server_theme.mp3")
	}

	s.initSpeaker()

	go s.playBackgroundMusic()

	fmt.Println("Server is running on http://localhost:", s.port)

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		fmt.Println("Error creating socket:", err)
		return
	}
	defer syscall.Close(fd)

	addr := syscall.SockaddrInet4{Port: s.port}
	copy(addr.Addr[:], []byte{127, 0, 0, 1})
	err = syscall.Bind(fd, &addr)
	if err != nil {
		fmt.Println("Error binding socket:", err)
		return
	}

	err = syscall.Listen(fd, 10)
	if err != nil {
		fmt.Println("Error listening on socket:", err)
		return
	}

	for {
		connFd, _, err := syscall.Accept(fd)
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go s.handleConnection(connFd)
	}
}

func (s *Server) initSpeaker() {
	speaker.Init(44100, 4410) // 44100 Hz sample rate, 4410 buffer size
	speaker.Play(s.mixer)     // Start the mixer on the speaker
}

func (s *Server) AddRoute(method, url string, handler func() string) {
	if s.routes[method] == nil {
		s.routes[method] = make(map[string]func() string)
	}
	s.routes[method][url] = handler
}

func (s *Server) AddSound(url string, soundFile string) {
	s.soundMap[url] = soundFile
}

func (s *Server) SetBackgroundMusic(musicFile string) {
	s.bgMusic = musicFile
}

func (s *Server) handleConnection(fd int) {
	defer syscall.Close(fd)

	buffer := make([]byte, 1024)
	n, err := syscall.Read(fd, buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	request := string(buffer[:n])
	fmt.Println("Request received:\n" + request)

	lines := strings.Split(request, "\r\n")
	if len(lines) > 0 {
		requestLine := strings.Fields(lines[0])
		if len(requestLine) > 1 {
			method := requestLine[0]
			url := requestLine[1]

			handler, exists := s.routes[method][url]
			if exists {
				response := s.createResponse(200, handler())
				s.writeResponse(fd, response)

				soundFile, ok := s.soundMap[url]
				if !ok {
					soundFile = filepath.Join(s.defaultDir, method+"_default_sound.mp3")
				}
				go s.playSound(soundFile)
			} else {
				response := s.createResponse(404, "Not Found")
				s.writeResponse(fd, response)

				// Play default sound for 404 errors
				default404Sound := filepath.Join(s.defaultDir, "404_sound.mp3")
				go s.playSound(default404Sound)
			}
		}
	}
}

func (s *Server) createResponse(statusCode int, body string) string {
	statusText := map[int]string{
		200: "OK",
		404: "Not Found",
		405: "Method Not Allowed",
	}[statusCode]

	return fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
		statusCode, statusText, len(body), body)
}

func (s *Server) writeResponse(fd int, response string) {
	_, err := syscall.Write(fd, []byte(response))
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

type VolumeStreamer struct {
	Streamer beep.Streamer
	Volume   float64 // A factor >1 increases volume, <1 decreases volume
}

// Stream modifies the samples to adjust volume.
func (v *VolumeStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = v.Streamer.Stream(samples)
	for i := range samples {
		samples[i][0] *= v.Volume // Left channel
		samples[i][1] *= v.Volume // Right channel
	}
	return n, ok
}

// Err propagates the error from the underlying streamer.
func (v *VolumeStreamer) Err() error {
	return v.Streamer.Err()
}

func (s *Server) playSound(soundFile string) {
	absPath, err := filepath.Abs(soundFile)
	if err != nil {
		fmt.Println("Error resolving sound file path:", err)
		return
	}

	fmt.Println("Playing sound file:", absPath)

	file, err := os.Open(absPath)
	if err != nil {
		fmt.Println("Error opening sound file:", err)
		return
	}

	streamer, _, err := mp3.Decode(file)
	if err != nil {
		fmt.Println("Error decoding sound file:", err)
		file.Close()
		return
	}

	volumeStreamer := &VolumeStreamer{
		Streamer: streamer,
		Volume:   5.0,
	}

	speaker.Lock()
	if s.mixer == nil {
		s.mixer = &beep.Mixer{}
	}
	s.mixer.Add(beep.Seq(volumeStreamer, beep.Callback(func() {
		file.Close()
	})))
	speaker.Unlock()
}

func (s *Server) playBackgroundMusic() {
	for {
		file, err := os.Open(s.bgMusic)
		if err != nil {
			fmt.Println("Error opening background music file:", err)
			return
		}
		defer file.Close()

		streamer, _, err := mp3.Decode(file)
		if err != nil {
			fmt.Println("Error decoding background music file:", err)
			return
		}
		defer streamer.Close()

		speaker.Lock()
		s.mixer.Add(beep.Loop(-1, streamer))
		speaker.Unlock()

		select {}
	}
}
