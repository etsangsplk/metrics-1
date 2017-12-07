// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScalarMetricDuplicates(t *testing.T) {
	r, _ := New()
	opts := Opts{
		Name: "foo",
		Help: "help",
	}
	_, err := r.NewCounter(opts)
	assert.NoError(t, err, "Failed first registration.")

	t.Run("same type", func(t *testing.T) {
		// You can't reuse options with the same metric type.
		_, err := r.NewCounter(opts)
		assert.Error(t, err)
	})

	t.Run("different type", func(t *testing.T) {
		// Even if you change the metric type, you still can't re-use metadata.
		_, err := r.NewGauge(opts)
		assert.Error(t, err)
	})

	t.Run("different help", func(t *testing.T) {
		// Changing the help string doesn't change the metric's identity.
		_, err := r.NewCounter(Opts{
			Name: "foo",
			Help: "different help",
		})
		assert.Error(t, err)
	})

	t.Run("added dimensions", func(t *testing.T) {
		// Can't have the same metric name with added dimensions.
		_, err := r.NewCounter(Opts{
			Name:   "foo",
			Help:   "help",
			Labels: Labels{"bar": "baz"},
		})
		assert.Error(t, err)
	})

	t.Run("different dimensions", func(t *testing.T) {
		// Even if the number of dimensions is the same, metrics with the same
		// name must have the same dimensions.
		_, err := r.NewCounter(Opts{
			Name:   "dimensions",
			Help:   "help",
			Labels: Labels{"bar": "baz"},
		})
		assert.NoError(t, err, "Failed to register new metric.")
		_, err = r.NewCounter(Opts{
			Name:   "dimensions",
			Help:   "help",
			Labels: Labels{"bing": "quux"},
		})
		assert.Error(t, err)
	})

	t.Run("same dimensions", func(t *testing.T) {
		// If a metric has the same name and dimensions, the label values may
		// change. This allows users to (inefficiently) create what are
		// effectively vectors - a collection of metrics with the same name and
		// label names, but different label values.
		_, err := r.NewCounter(Opts{
			Name:   "dimensions",
			Help:   "help",
			Labels: Labels{"bar": "quux"},
		})
		assert.NoError(t, err)
	})

	t.Run("duplicate scrubbed name", func(t *testing.T) {
		// Uniqueness is enforced after the metric name is scrubbed.
		_, err := r.NewCounter(Opts{
			Name: "scrubbed_name",
			Help: "help",
		})
		assert.NoError(t, err, "Failed to register new metric.")
		_, err = r.NewCounter(Opts{
			Name: "scrubbed&name",
			Help: "help",
		})
		assert.Error(t, err)
	})

	t.Run("duplicate scrubbed dimensions", func(t *testing.T) {
		// Uniqueness is enforced after labels are scrubbed.
		_, err := r.NewCounter(Opts{
			Name:   "scrubbed_dimensions",
			Help:   "help",
			Labels: Labels{"b_r": "baz"},
		})
		assert.NoError(t, err, "Failed to register new metric.")
		_, err = r.NewCounter(Opts{
			Name:   "scrubbed_dimensions",
			Help:   "help",
			Labels: Labels{"b&r": "baz"},
		})
		assert.Error(t, err)
	})

	t.Run("constant label name specified twice", func(t *testing.T) {
		// Within a single user-supplied set of labels, scrubbing may not
		// introduce duplicates.
		_, err = r.NewCounter(Opts{
			Name:   "user_error_constant_labels",
			Help:   "help",
			Labels: Labels{"b_r": "baz", "b&r": "baz"},
		})
		assert.Error(t, err)
	})
}

func TestVectorMetricDuplicates(t *testing.T) {
	r, _ := New()
	opts := Opts{
		Name:           "foo",
		Help:           "help",
		VariableLabels: []string{"foo"},
	}
	_, err := r.NewCounterVector(opts)
	assert.NoError(t, err, "Failed first registration.")

	t.Run("same type", func(t *testing.T) {
		// You can't reuse options with the same metric type.
		_, err := r.NewCounterVector(opts)
		assert.Error(t, err, "Unexpected success re-using vector metrics metadata.")
	})

	t.Run("different type", func(t *testing.T) {
		// Even if you change the metric type, you still can't re-use metadata.
		_, err := r.NewGaugeVector(opts)
		assert.Error(t, err, "Unexpected success re-using vector metrics metadata.")
	})

	t.Run("different help", func(t *testing.T) {
		// Changing the help string doesn't change the metric's identity.
		_, err := r.NewCounterVector(Opts{
			Name:           "foo",
			Help:           "different help",
			VariableLabels: []string{"foo"},
		})
		assert.Error(t, err)
	})

	t.Run("added dimensions", func(t *testing.T) {
		// Can't have the same metric name with added dimensions.
		_, err := r.NewCounterVector(Opts{
			Name:           "foo",
			Help:           "help",
			VariableLabels: []string{"foo"},
			Labels:         Labels{"bar": "baz"},
		})
		assert.Error(t, err, "Shouldn't be able to add constant labels.")
		_, err = r.NewCounterVector(Opts{
			Name:           "foo",
			Help:           "help",
			VariableLabels: []string{"foo", "bar"},
		})
		assert.Error(t, err, "Shouldn't be able to add variable labels.")
	})

	t.Run("different dimensions", func(t *testing.T) {
		// Even if the number of dimensions is the same, metrics with the same
		// name must have the same dimensions.
		_, err := r.NewCounterVector(Opts{
			Name:           "foo",
			Help:           "help",
			VariableLabels: []string{"bar"},
		})
		assert.Error(t, err)
	})

	t.Run("same dimensions", func(t *testing.T) {
		// If a metric has the same name and dimensions, the label values
		// may change. (Again, this would be more efficiently modeled as a
		// higher-dimensionality vector.)
		_, err := r.NewCounterVector(Opts{
			Name:           "dimensions",
			Help:           "help",
			Labels:         Labels{"bar": "baz"},
			VariableLabels: []string{"foo"},
		})
		assert.NoError(t, err)
		_, err = r.NewCounterVector(Opts{
			Name:           "dimensions",
			Help:           "help",
			Labels:         Labels{"bar": "quux"},
			VariableLabels: []string{"foo"},
		})
		assert.NoError(t, err)
	})

	t.Run("vectors own dimensions", func(t *testing.T) {
		// If a vector with given dimensions exists, scalars that could be part of
		// that vector may not exist. In other words, for a given set of
		// dimensions, users can't sometimes use a vector and sometimes use a la
		// carte scalars.

		// dims: foo, baz
		_, err := r.NewCounterVector(Opts{
			Name:           "ownership",
			Help:           "help",
			Labels:         Labels{"foo": "bar"},
			VariableLabels: []string{"baz"},
		})
		require.NoError(t, err)

		// same dims
		_, err = r.NewCounter(Opts{
			Name:   "ownership",
			Help:   "help",
			Labels: Labels{"foo": "bar", "baz": "quux"},
		})
		require.Error(t, err)
	})

	t.Run("duplicate scrubbed name", func(t *testing.T) {
		// Uniqueness is enforced after the metric name is scrubbed.
		_, err := r.NewCounterVector(Opts{
			Name:           "scrubbed_name",
			Help:           "help",
			VariableLabels: []string{"bar"},
		})
		assert.NoError(t, err, "Failed to register new metric.")
		_, err = r.NewCounterVector(Opts{
			Name:           "scrubbed&name",
			Help:           "help",
			VariableLabels: []string{"bar"},
		})
		assert.Error(t, err)
	})

	t.Run("duplicate scrubbed dimensions", func(t *testing.T) {
		// Uniqueness is enforced after labels are scrubbed.
		_, err := r.NewCounterVector(Opts{
			Name:           "scrubbed_dimensions",
			Help:           "help",
			Labels:         Labels{"b_r": "baz"},
			VariableLabels: []string{"q__x"},
		})
		assert.NoError(t, err, "Failed to register new metric.")
		_, err = r.NewCounterVector(Opts{
			Name:           "scrubbed_dimensions",
			Help:           "help",
			Labels:         Labels{"b&r": "baz"},
			VariableLabels: []string{"q&&x"},
		})
		assert.Error(t, err)
	})

	t.Run("constant label name specified twice", func(t *testing.T) {
		// Within a single user-supplied set of constant labels, scrubbing may not
		// introduce duplicates.
		_, err = r.NewCounterVector(Opts{
			Name:           "user_error_constant_labels",
			Help:           "help",
			Labels:         Labels{"b_r": "baz", "b&r": "baz"},
			VariableLabels: []string{"quux"},
		})
		assert.Error(t, err)
	})

	t.Run("variable label name specified twice", func(t *testing.T) {
		// Within a single user-supplied set of variable labels, scrubbing may not
		// introduce duplicates.
		_, err = r.NewCounterVector(Opts{
			Name:           "user_error_variable_labels",
			Help:           "help",
			VariableLabels: []string{"f__", "f&&"},
		})
		assert.Error(t, err)
	})
}