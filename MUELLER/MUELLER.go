package MUELLER

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type Server struct {
	port           int
	routes         map[string]map[string]func() string
	soundMap       map[string]string
	bgMusic        string
	defaultDir     string
	mixer          *beep.Mixer
	powerVerified  bool
	artCollections int
}

var cables = map[string]struct {
	connector string
	voltage   int
	wattage   int
}{
	"usbc":      {"usbc", 5, 60},         // Mueller's special 60W cable
	"usba":      {"usba", 5, 5},          // Regular USBA
	"powerjack": {"powerjack", 220, 500}, // Dangerous!
	"hdmi":      {"hdmi", 0, 0},          // For the lulz
}

var adapters = map[string]struct {
	from       string
	to         string
	voltageIn  int
	voltageOut int
	memeText   string
}{
	"to-usba":      {"usbc", "usba", 5, 5, "🔌 Adapter approved by ThinkPad consortium 🔌"},
	"to-hdmi":      {"usba", "hdmi", 5, 0, "🎮 Now displaying on 1997 CRT monitor 🎮"},
	"step-up":      {"usbc", "usbc", 5, 220, "⚡ DANGER! Mueller's homemade adapter ⚡"},
	"to-powerjack": {"usba", "powerjack", 5, 220, "🔥 FIRE HAZARD! But art requires risks 🔥"},
}

type component struct {
	compType string
	value    string
}

type RouteBuilder struct {
	server    *Server
	method    string
	path      string
	chain     []component
	isArt     bool
	powerWatt int
	powerType string
}

func NewServer(port int) *Server {
	rand.Seed(time.Now().UnixNano())
	return &Server{
		port:       port,
		routes:     make(map[string]map[string]func() string),
		soundMap:   make(map[string]string),
		defaultDir: "server_sounds",
		mixer:      &beep.Mixer{},
	}
}

func (s *Server) Route(method, path string) *RouteBuilder {
	fmt.Printf("📡 Routing %s %s through Linux kernel v%d.%d+ 📡\n", method, path, rand.Intn(6), rand.Intn(30))
	return &RouteBuilder{
		server:    s,
		method:    method,
		path:      path,
		powerWatt: -1,
	}
}

func (rb *RouteBuilder) Cable(cableType string) *RouteBuilder {
	rb.chain = append(rb.chain, component{"cable", cableType})
	fmt.Printf("🔌 Added %s cable (resistance is futile) 🔌\n", cableType)
	return rb
}

func (rb *RouteBuilder) Adapter(adapterType string) *RouteBuilder {
	rb.chain = append(rb.chain, component{"adapter", adapterType})
	if a, ok := adapters[adapterType]; ok {
		fmt.Println(a.memeText)
	}
	return rb
}

func (rb *RouteBuilder) PowerSupply(wattage int, cableType string) *RouteBuilder {
	rb.powerWatt = wattage
	rb.powerType = cableType
	fmt.Printf("⚡ Power check: %dW via %s ⚡\n", wattage, cableType)
	return rb
}

func Art(handler func() string) func() string {
	fmt.Println("🎨 Someone asked if this is art! 🎨")
	return func() string {
		return "🏛️ Museum of Modern Code 🏛️\n" + handler() + "\n🏛️ Curated by Mueller 🏛️"
	}
}

func (rb *RouteBuilder) Handler(handler func() string) {
	if !rb.server.powerVerified {
		if rb.powerWatt != 60 || rb.powerType != "usbc" {
			panic("\n💥 NO POWER! Mueller's ThinkPad requires exactly 60W via USBC! 💥\n(He's playing Plants vs Zombies in the background)\n")
		}
		rb.server.powerVerified = true
		fmt.Println("⚡ 60W Power Verified! ThinkPad whirrs to life ⚡")
	}

	if err := rb.validateChain(); err != nil {
		panic(fmt.Sprintf("\n💥 Mueller Framework Error 💥\n%s\n", err))
	}

	if !rb.isArt {
		fmt.Printf("⚠️ WARNING: Route %s not marked as art! Mueller's disappointment is immeasurable. ⚠️\n", rb.path)
	} else {
		rb.server.artCollections++
		fmt.Printf("🎨 Art detected! Added to Mueller's museum (Total: %d) 🎨\n", rb.server.artCollections)
	}

	rb.server.AddRoute(rb.method, rb.path, handler)
}

func (rb *RouteBuilder) validateChain() error {
	if len(rb.chain) == 0 {
		return fmt.Errorf("🛑 Route must start with a cable! You can't connect thin air! 🌌")
	}

	firstComp := rb.chain[0]
	if firstComp.compType != "cable" {
		return fmt.Errorf("🚫 First component must be a cable! You started with %s. 🚫", firstComp.compType)
	}

	firstCable, ok := cables[firstComp.value]
	if !ok {
		return fmt.Errorf("🚫 What's a %s cable? You making this up? 🚫", firstComp.value)
	}

	currentType := firstCable.connector
	currentVoltage := firstCable.voltage

	for i := 1; i < len(rb.chain); i++ {
		comp := rb.chain[i]
		switch comp.compType {
		case "cable":
			cable, ok := cables[comp.value]
			if !ok {
				return fmt.Errorf("🚫 Unknown cable: %s. Did you find this in a time capsule? 🚫", comp.value)
			}
			if cable.connector != currentType {
				return fmt.Errorf("💥 KA-BOOM! Can't connect %s to %s without an adapter! 💥", currentType, cable.connector)
			}
			if cable.voltage != currentVoltage {
				return fmt.Errorf("🔥 %dV into %dV? You'll fry the server! 🔥 (Voltage mismatch)", currentVoltage, cable.voltage)
			}

		case "adapter":
			adapter, ok := adapters[comp.value]
			if !ok {
				return fmt.Errorf("🚫 Adapter %s not found. Check your grandma's basement! 🚫", comp.value)
			}
			if adapter.from != currentType {
				return fmt.Errorf("🔌 Adapter %s needs %s input, but got %s. 🔌", comp.value, adapter.from, currentType)
			}
			if adapter.voltageIn != currentVoltage {
				return fmt.Errorf("⚡️ ZAP! %s requires %dV, but got %dV. ⚡️", comp.value, adapter.voltageIn, currentVoltage)
			}
			currentType = adapter.to
			currentVoltage = adapter.voltageOut

			if currentVoltage >= 220 {
				fmt.Printf("⚠️ WARNING: %dV! Wear rubber gloves! ⚠️\n", currentVoltage)
			}

		default:
			return fmt.Errorf("🚫 Unknown component: %s. 🚫", comp.compType)
		}
	}

	if len(rb.chain) > 3 {
		fmt.Println("🕸️ Whoa, that's a rats nest of adapters! 🕸️")
	}

	// Fish people detection
	if rand.Intn(100) < 30 {
		fmt.Println("🐠 Mueller thinks someone in the room looks like a fish 🐠")
	}

	return nil
}

func (s *Server) Start() {
	fmt.Println("\n   *                        (        (                (     \n (  `                       )\\ )     )\\ )             )\\ )  \n )\\))(        (     (      (()/(    (()/(     (      (()/(  \n((_)()\\       )\\    )\\      /(_))    /(_))    )\\      /(_)) \n(_()((_)   _ ((_)  ((_)    (_))     (_))     ((_)    (_))   \n|  \\/  |  | | | |  | __|   | |      | |      | __|   | _ \\  \n| |\\/| | _| |_| |_ | _|  _ | |__  _ | |__  _ | _|  _ |   /  \n|_|  |_|(_)\\___/(_)|___|(_)|____|(_)|____|(_)|___|(_)|_|_\\\n")
	fmt.Println("The kleine süße Lösung is starting: \n")
	fmt.Println("🐧 Initializing Linux subsystem...")
	fmt.Println("🖱️ Calibrating ThinkPad trackpoint...")
	fmt.Println("🌱 Loading Plants vs Zombies defense system...")

	if s.bgMusic == "" {
		s.bgMusic = filepath.Join(s.defaultDir, "server_theme.mp3")
	}

	s.initSpeaker()

	go s.playBackgroundMusic()

	fmt.Printf("\n🚀 Server launching on http://localhost:%d 🚀\n", s.port)
	fmt.Println("⚠️ Warning: Mueller is judging your cable management ⚠️")

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
	speaker.Init(44100, 4410)
	speaker.Play(s.mixer)
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
	if strings.Contains(request, "Zombie") {
		fmt.Println("🧟 Zombie detected! Launching lawnmower... 🧹")
	}

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
	Volume   float64
}

func (v *VolumeStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = v.Streamer.Stream(samples)
	for i := range samples {
		samples[i][0] *= v.Volume
		samples[i][1] *= v.Volume
	}
	return n, ok
}

func (v *VolumeStreamer) Err() error {
	return v.Streamer.Err()
}

func (s *Server) playSound(soundFile string) {
	pvzSounds := map[string]string{
		"GET_default_sound.mp3":    "pvz_plant.mp3",
		"POST_default_sound.mp3":   "pvz_shovel.mp3",
		"404_sound.mp3":            "pvz_zombie.mp3",
		"DELETE_default_sound.mp3": "pvz_scream.mp3",
	}

	if newSound, ok := pvzSounds[filepath.Base(soundFile)]; ok {
		soundFile = filepath.Join(s.defaultDir, newSound)
	}

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
