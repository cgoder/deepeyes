package gss

// func gss() {

// 	// go http.ListenAndServe("0.0.0.0:9527", nil)

// 	log.WithFields(logrus.Fields{
// 		"module": "main",
// 		"func":   "main",
// 	}).Info("GSS Server start")

// 	_, cancel := context.WithCancel(context.Background())
// 	defer func() {
// 		cancel()
// 	}()

// 	svr := NewService()
// 	// log.SetLevel(svr.ServerLogLevel())
// 	go api.HTTPAPIServer(svr)
// 	go DebugRuntime()
// 	// go RTSPServer()
// 	// go Storage.StreamChannelRunAll(ctx)

// 	signalChanel := make(chan os.Signal, 1)
// 	done := make(chan bool, 1)
// 	signal.Notify(signalChanel, syscall.SIGINT, syscall.SIGTERM)
// 	go func() {
// 		sig := <-signalChanel
// 		log.WithFields(logrus.Fields{
// 			"module": "main",
// 			"func":   "main",
// 		}).Info("Server receive signal", sig)
// 		done <- true
// 	}()
// 	log.WithFields(logrus.Fields{
// 		"module": "main",
// 		"func":   "main",
// 	}).Info("Server start success a wait signals")

// 	<-done

// 	svr.ProgramStopAll()

// 	time.Sleep(2 * time.Second)

// 	log.WithFields(logrus.Fields{
// 		"module": "main",
// 		"func":   "main",
// 	}).Info("Server stop working by signal")
// }
