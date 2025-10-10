package l1

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	config "github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"

	_ "embed"
)

//go:embed params.yaml
var params []byte

const (
	enclaveName         = "localnet"
	kurtosisPackageName = "github.com/ssvlabs/ssv-mini"
)

func start(ctx context.Context) error {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return errors.Join(err, errors.New("failed to create kurtosis context"))
	}

	enclaveCtx, err := kurtosisCtx.CreateEnclave(ctx, enclaveName)
	if err != nil {
		return errors.Join(err, errors.New("failed to create enclave"))
	}

	outputCh, cancel, err := enclaveCtx.RunStarlarkRemotePackage(
		ctx,
		kurtosisPackageName,
		config.NewRunStarlarkConfig(config.WithSerializedParams(string(params))))
	if err != nil {
		return errors.Join(err, errors.New("failed to run starlark package"))
	}
	defer cancel()

	var jsonResponse string

	for output := range outputCh {
		slog.Debug(output.String())

		info := output.GetInfo()
		if info != nil {
			slog.Info(info.GetInfoMessage())
		}

		if pr := output.GetProgressInfo(); pr != nil {
			for _, info := range pr.GetCurrentStepInfo() {
				slog.Info(info,
					slog.Any("step_number", pr.GetCurrentStepNumber()),
					slog.Any("total_steps", pr.GetTotalSteps()),
				)
			}
		}
		if i := output.GetInstruction(); i != nil {
			slog.Info("executing instruction", slog.Any("instruction", i.ExecutableInstruction))
		}

		if kurtosisWarn := output.GetWarning(); kurtosisWarn != nil {
			slog.Warn(kurtosisWarn.String())
		}

		kurtosisErr := output.GetError()
		if kurtosisErr != nil {
			slog.Error(kurtosisErr.String())
			return fmt.Errorf("kurtosis package returned error: %s", kurtosisErr.String())
		}

		ev := output.GetRunFinishedEvent()
		if ev != nil && ev.SerializedOutput != nil {
			jsonResponse = *ev.SerializedOutput
		}
	}

	slog.Info("Kurtosis package launched",
		slog.String("package", kurtosisPackageName),
		slog.String("response", jsonResponse))

	return nil
}
