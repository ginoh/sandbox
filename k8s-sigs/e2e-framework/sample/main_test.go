package sample_test

import (
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
)

var (
	testEnv env.Environment
)

func TestMain(m *testing.M) {
	//cfg, _ := envconf.NewFromFlags()
	//testEnv = env.NewWithConfig(cfg)
	testEnv, _ = env.NewFromFlags()
	kindClusterName := envconf.RandomName("sample-cluster", 16)
	namespace := envconf.RandomName("sample-ns", 16)

	// Use pre-defined environment funcs to create a kind cluster prior to test run
	testEnv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
		envfuncs.CreateNamespace(namespace),
	)

	// Use pre-defined environment funcs to teardown kind cluster after tests
	testEnv.Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.DestroyCluster(kindClusterName),
	)

	// launch package tests
	os.Exit(testEnv.Run(m))
}
