package mx

import "github.com/sudachen/go-ml/nn/mx/capi"

func SgdMomUpdate(params, grads, mom *NDArray, lr, momentum, wd float32) {
	capi.OptimizerUpdate(
		capi.OpSgdMomUpdate,
		params.handle, grads.handle, mom.handle, nil,
		capi.KeyLr, lr,
		capi.KeyMomentum, momentum,
		capi.KeyWd, wd)
}

func SgdUpdate(params, grads *NDArray, lr, wd float32)  {
	capi.OptimizerUpdate(
		capi.OpSgdUpdate,
		params.handle, grads.handle, nil, nil,
		capi.KeyLr, lr,
		capi.KeyWd, wd)
}

func AdamUpdate(params, grads, mean, variance *NDArray, lr, beta1, beta2, epsilon, wd float32) {
	capi.OptimizerUpdate(
		capi.OpAdamUpdate,
		params.handle, grads.handle, mean.handle, variance.handle,
		capi.KeyLr, lr,
		capi.KeyBeta1, beta1,
		capi.KeyBeta2, beta2,
		capi.KeyEpsilon, epsilon,
		capi.KeyWd, wd)
}