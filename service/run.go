package service

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coldze/primitives/custom_error"
)

type MainFunc func(stopping <-chan struct{}) int

type gracefulShutdown struct {
	WaitForShutdown  <-chan struct{}
	ShutdownComplete chan<- int
	ReturnCode       <-chan int
}

func runGracefully(timeout time.Duration) *gracefulShutdown {
	shutdown := make(chan struct{}, 10)
	shutdownComplete := make(chan int, 10)
	returnCode := make(chan int, 10)

	gracefulStop := make(chan os.Signal, 10)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		defer close(shutdown)
		defer close(gracefulStop)
		exitCode := 0
		select {
		case exitCode = <-shutdownComplete:
			{
				log.Printf("Business logic completed. Exit code: %+v", exitCode)
				returnCode <- exitCode
				return
			}

		case sig := <-gracefulStop:
			log.Printf("Caught sig: %+v", sig)
			shutdown <- struct{}{}
		}

		select {
		case <-time.After(timeout):
			log.Print("Shutdown timeout occured. Terminating.")
			returnCode <- 1
		case exitCode := <-shutdownComplete:
			log.Print("Shutdown complete.")
			returnCode <- exitCode
		}
	}()

	return &gracefulShutdown{
		WaitForShutdown:  shutdown,
		ShutdownComplete: shutdownComplete,
		ReturnCode:       returnCode,
	}
}

func safeRunAppLogic(appLogic MainFunc, stopChan <-chan struct{}) (res int) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		customErr, ok := r.(custom_error.CustomError)
		if ok {
			log.Printf("mainFunc failed. Error: %v", customErr)
			res = 1
			return
		}
		err, ok := r.(error)
		if ok {
			log.Printf("mainFunc failed. Error: %v", err)
			res = 1
			return
		}
		log.Printf("mainFunc failed. Unknown error: %+v. Type: %T", r, r)
		res = 1
	}()
	return appLogic(stopChan)
}

func Run(timeout time.Duration, appLogic MainFunc) {

	graceful := runGracefully(timeout)
	go func() {
		graceful.ShutdownComplete <- safeRunAppLogic(appLogic, graceful.WaitForShutdown)
	}()
	exitCode := <-graceful.ReturnCode
	log.Printf("Exiting application. Code: %+v", exitCode)
	os.Exit(exitCode)
}
