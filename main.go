package main

import (
	"MUELLER/MUELLER"
	"fmt"
)

// Example handler that does not count as art
func normalHandler() string {
	return "Hello from a plain old route!"
}

// Example handler that is wrapped in the "Art" function
func fancyHandler() string {
	return "Observe the minimalistic synergy of cables and code."
}

func main() {
	server := MUELLER.NewServer(2000)

	// 1) A "normal" route: ensures 60W via USBC, then we do usbc->to-usba->usba
	//    so that it won't panic, we connect it properly and supply correct power
	server.
		Route("GET", "/students").
		Cable("usbc").           // Start with a valid cable
		Adapter("to-usba").      // Convert from usbc -> usba
		Cable("usba").           // We’re back at usba now
		PowerSupply(60, "usbc"). // Must be 60W via usbc or else it panics
		Handler(func() string {
			return "GET request received: students"
		})

	// 2) A "POST" route that is "art", with the same safe cable chain.
	//    Notice we wrap the handler in MUELLER.Art() to get the museum text.
	server.
		Route("POST", "/gallery").
		Cable("usbc").
		Adapter("to-usba").
		Cable("usba").
		PowerSupply(60, "usbc").
		Handler(MUELLER.Art(func() string {
			return "POST request received: gallery"
		}))

	// 3) A route that uses a "step-up" adapter to jump from 5V to 220V. This is "art" too.
	server.
		Route("GET", "/hautevoltage").
		Cable("usbc").
		Adapter("step-up"). // In adapters: from usbc to usbc, 5V -> 220V
		Cable("usbc").      // We come out the adapter still as usbc (but 220V now!)
		PowerSupply(60, "usbc").
		Handler(MUELLER.Art(func() string {
			return "We have stepped up the voltage to 220! Danger, high voltage!"
		}))

	// 4) Another normal route example—just to show we can still add more.
	server.
		Route("DELETE", "/students").
		Cable("usbc").
		Adapter("to-usba").
		Cable("usba").
		PowerSupply(60, "usbc").
		Handler(func() string {
			return "DELETE request received: students"
		})

	server.
		Route("GET", "/panicExample").
		Adapter("to-usba"). // No cable first -> immediate mismatch
		Cable("usba").
		PowerSupply(60, "usbc").
		Handler(func() string {
			return "If you see this, the Universe has broken"
		})

	fmt.Println("Starting Mueller Meme Server on port 2000...")
	server.Start()
}
