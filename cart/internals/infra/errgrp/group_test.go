package errgrp

import (
	"errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"math/rand"
	"testing"
	"time"
)

func asyncSuccessfulTask() error {
	time.Sleep(time.Duration(rand.Float32()+1.0) * time.Second)
	return nil
}

func asyncFailedTask(err error) error {
	time.Sleep(time.Duration(rand.Float32()+1) * time.Second)
	return err
}

func TestGr_NoErrors(t *testing.T) {
	defer goleak.VerifyNone(t)

	var g ErrGr

	g.Go(func() (err error) {
		return asyncSuccessfulTask()
	})

	g.Go(func() (err error) {
		return asyncSuccessfulTask()
	})

	g.Go(func() (err error) {
		return asyncSuccessfulTask()
	})

	if err := g.Wait(); err != nil {
		require.Fail(t, "")
		return
	}
}

func TestGr_HaveErrors(t *testing.T) {
	defer goleak.VerifyNone(t)

	var g ErrGr
	wantErr := errors.New("error")

	g.Go(func() (err error) {
		return asyncFailedTask(wantErr)
	})

	g.Go(func() (err error) {
		return asyncSuccessfulTask()
	})

	g.Go(func() (err error) {
		return asyncSuccessfulTask()
	})

	if err := g.Wait(); err != nil {
		require.ErrorIs(t, err, wantErr)
		return
	}

	require.Fail(t, "No errors")
}
