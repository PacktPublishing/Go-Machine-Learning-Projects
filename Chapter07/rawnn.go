package main

import (
	"math"

	"github.com/pkg/errors"
	"gonum.org/v1/gonum/stat/distuv"
	"gorgonia.org/tensor"
)

type NN struct {
	hidden, final *tensor.Dense
	b0, b1        float64
}

func New(input, hidden, output int) (retVal *NN) {
	r := make([]float64, hidden*input)
	r2 := make([]float64, hidden*output)
	fillRandom(r, float64(len(r)))
	fillRandom(r2, float64(len(r2)))
	hiddenT := tensor.New(tensor.WithShape(hidden, input), tensor.WithBacking(r))
	finalT := tensor.New(tensor.WithShape(output, hidden), tensor.WithBacking(r2))
	return &NN{
		hidden: hiddenT,
		final:  finalT,
	}
}

func (nn *NN) Predict(a tensor.Tensor) (int, error) {
	if a.Dims() != 1 {
		return -1, errors.New("Expected a vector")
	}

	var m maybe
	hidden := m.do(func() (tensor.Tensor, error) { return nn.hidden.MatVecMul(a) })
	act0 := m.do(func() (tensor.Tensor, error) { return hidden.Apply(sigmoid, tensor.UseUnsafe()) })

	final := m.do(func() (tensor.Tensor, error) { return tensor.MatVecMul(nn.final, act0) })
	pred := m.do(func() (tensor.Tensor, error) { return final.Apply(sigmoid, tensor.UseUnsafe()) })

	if m.err != nil {
		return -1, m.err
	}
	return argmax(pred.Data().([]float64)), nil
}

func (nn *NN) PredHid(a tensor.Tensor) (act0, pred tensor.Tensor, retVal int) {
	var m maybe
	hidden := m.do(func() (tensor.Tensor, error) { return nn.hidden.MatVecMul(a) })
	act0 = m.do(func() (tensor.Tensor, error) { return hidden.Apply(sigmoid, tensor.UseUnsafe()) })

	final := m.do(func() (tensor.Tensor, error) { return tensor.MatVecMul(nn.final, act0) })
	pred = m.do(func() (tensor.Tensor, error) { return final.Apply(sigmoid, tensor.UseUnsafe()) })

	if m.err != nil {
		return nil, nil, 0
	}
	retVal = argmax(pred.Data().([]float64))
	return
}

// X is the image, Y is a one hot vector
func (nn *NN) Train(x, y tensor.Tensor, learnRate float64) (cost float64, err error) {
	// predict
	var m maybe
	m.do(func() (tensor.Tensor, error) { err := x.Reshape(x.Shape()[0], 1); return x, err })
	m.do(func() (tensor.Tensor, error) { err := y.Reshape(10, 1); return y, err })

	hidden := m.do(func() (tensor.Tensor, error) { return tensor.MatMul(nn.hidden, x) })
	act0 := m.do(func() (tensor.Tensor, error) { return hidden.Apply(sigmoid, tensor.UseUnsafe()) })

	final := m.do(func() (tensor.Tensor, error) { return tensor.MatMul(nn.final, act0) })
	pred := m.do(func() (tensor.Tensor, error) { return final.Apply(sigmoid, tensor.UseUnsafe()) })
	// log.Printf("pred %v, correct %v", argmax(pred.Data().([]float64)), argmax(y.Data().([]float64)))

	// backpropagation.
	outputErrors := m.do(func() (tensor.Tensor, error) { return tensor.Sub(y, pred) })
	cost = sum(outputErrors.Data().([]float64))

	hidErrs := m.do(func() (tensor.Tensor, error) {
		if err := nn.final.T(); err != nil {
			return nil, err
		}
		defer nn.final.UT()
		return tensor.MatMul(nn.final, outputErrors)
	})

	if m.err != nil {
		return 0, m.err
	}

	dpred := m.do(func() (tensor.Tensor, error) { return pred.Apply(dsigmoid, tensor.UseUnsafe()) })
	m.do(func() (tensor.Tensor, error) { return tensor.Mul(dpred, outputErrors, tensor.UseUnsafe()) })
	// m.do(func() (tensor.Tensor, error) { err := act0.T(); return act0, err })
	dpred_dfinal := m.do(func() (tensor.Tensor, error) {
		if err := act0.T(); err != nil {
			return nil, err
		}
		defer act0.UT()
		return tensor.MatMul(outputErrors, act0)
	})

	dact0 := m.do(func() (tensor.Tensor, error) { return act0.Apply(dsigmoid) })
	m.do(func() (tensor.Tensor, error) { return tensor.Mul(hidErrs, dact0, tensor.UseUnsafe()) })
	m.do(func() (tensor.Tensor, error) { err := hidErrs.Reshape(hidErrs.Shape()[0], 1); return hidErrs, err })
	// m.do(func() (tensor.Tensor, error) { err := x.T(); return x, err })
	dcost_dhidden := m.do(func() (tensor.Tensor, error) {
		if err := x.T(); err != nil {
			return nil, err
		}
		defer x.UT()
		return tensor.MatMul(hidErrs, x)
	})

	// gradient update
	m.do(func() (tensor.Tensor, error) { return tensor.Mul(dpred_dfinal, learnRate, tensor.UseUnsafe()) })
	m.do(func() (tensor.Tensor, error) { return tensor.Mul(dcost_dhidden, learnRate, tensor.UseUnsafe()) })
	m.do(func() (tensor.Tensor, error) { return tensor.Add(nn.final, dpred_dfinal, tensor.UseUnsafe()) })
	m.do(func() (tensor.Tensor, error) { return tensor.Add(nn.hidden, dcost_dhidden, tensor.UseUnsafe()) })
	return cost, m.err
}

func sigmoid(a float64) float64 { return 1 / (1 + math.Exp(-1*a)) }

func dsigmoid(a float64) float64 { return (1 - a) * a }

func onesLike(t tensor.Tensor) tensor.Tensor {
	retVal := t.Clone().(tensor.Tensor)
	data := retVal.Data().([]float64)
	for i := range data {
		data[i] = 1
	}
	return retVal
}

func fillRandom(a []float64, v float64) {
	dist := distuv.Uniform{
		Min: -1 / math.Sqrt(v),
		Max: 1 / math.Sqrt(v),
	}
	for i := range a {
		a[i] = dist.Rand()
	}
}

type maybe struct {
	err error
}

func (m *maybe) do(fn func() (tensor.Tensor, error)) tensor.Tensor {
	if m.err != nil {
		return nil
	}

	var retVal tensor.Tensor
	if retVal, m.err = fn(); m.err == nil {
		return retVal
	}
	m.err = errors.WithStack(m.err)
	return nil
}
