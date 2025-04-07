package main

import (
	"os"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

func main(){

  cnfg := ctrl.GetConfigOrDie()

  mgr,err := ctrl.NewManager(cnfg,ctrl.Options{})

  if err != nil {
    setupLog.Error(err, "Unable to create the manager")
    os.Exit(1)
  }

  setupLog.Info("manager is created ", mgr)

}
