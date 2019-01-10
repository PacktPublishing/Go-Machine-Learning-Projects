package main

import (
	"math"

	"gorgonia.org/tensor"
	"gorgonia.org/tensor/native"
	"gorgonia.org/vecf64"
)

func zca(data tensor.Tensor) (retVal tensor.Tensor, err error) {
	var dataᵀ, data2, sigma tensor.Tensor
	data2 = data.Clone().(tensor.Tensor)

	if err := minusMean(data2); err != nil {
		return nil, err
	}
	if dataᵀ, err = tensor.T(data2); err != nil {
		return nil, err
	}

	if sigma, err = tensor.MatMul(dataᵀ, data2); err != nil {
		return nil, err
	}

	cols := sigma.Shape()[1]
	if _, err = tensor.Div(sigma, float64(cols), tensor.UseUnsafe()); err != nil {
		return nil, err
	}

	s, u, _, err := sigma.(*tensor.Dense).SVD(true, true)
	if err != nil {
		return nil, err
	}

	var diag, uᵀ, tmp tensor.Tensor
	if diag, err = s.Apply(invSqrt(0.08), tensor.UseUnsafe()); err != nil {
		return nil, err
	}
	diag = tensor.New(tensor.AsDenseDiag(diag))

	if uᵀ, err = tensor.T(u); err != nil {
		return nil, err
	}

	if tmp, err = tensor.MatMul(u, diag); err != nil {
		return nil, err
	}

	if tmp, err = tensor.MatMul(tmp, uᵀ); err != nil {
		return nil, err
	}

	if err = tmp.T(); err != nil {
		return nil, err
	}

	return tensor.MatMul(data, tmp)
}

func invSqrt(epsilon float64) func(float64) float64 {
	return func(a float64) float64 {
		return 1 / math.Sqrt(a+epsilon)
	}

}

func minusMean(a tensor.Tensor) error {
	nat, err := native.MatrixF64(a.(*tensor.Dense))
	if err != nil {
		return err
	}
	for _, row := range nat {
		mean := avg(row)
		vecf64.Trans(row, -mean)

		// standardization
		// var stdev float64
		// for _, col := range row {
		// 	stdev += (col - mean) * (col - mean)
		// }
		// stdev /= float64(len(row))
		// stdev = math.Sqrt(stdev)

		// for j := range row {
		// 	row[j] = (row[j] - mean) / stdev
		// }
	}

	rows, cols := a.Shape()[0], a.Shape()[1]

	mean := make([]float64, cols)
	for j := 0; j < cols; j++ {
		var colMean float64
		for i := 0; i < rows; i++ {
			colMean += nat[i][j]
		}
		colMean /= float64(rows)
		mean[j] = colMean
	}

	for _, row := range nat {
		vecf64.Sub(row, mean)
	}

	return nil
}
