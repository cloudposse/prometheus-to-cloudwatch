package main

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/common/model"
)

func Test_getName(t *testing.T) {
	cases := map[string]struct {
		m        model.Metric
		expected string
	}{
		"has_name_label":           {model.Metric{model.MetricNameLabel: "foo"}, "foo"},
		"does_not_have_name_label": {model.Metric{"foo": "bar"}, ""},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			if n := getName(c.m); n != c.expected {
				t.Errorf("got %q; expected %q", n, c.expected)
			}
		})
	}
}

func Test_getDimensions(t *testing.T) {
	cases := map[string]struct {
		m        model.Metric
		expected []*cloudwatch.Dimension
	}{
		"no_labels":       {model.Metric{}, []*cloudwatch.Dimension{}},
		"only_name_label": {model.Metric{model.MetricNameLabel: "foo"}, []*cloudwatch.Dimension{}},
		"less_than_10_labels": {
			model.Metric{model.MetricNameLabel: "foo", "host": "prod-host01", "vpc": "pub-us-east-1d"},
			[]*cloudwatch.Dimension{new(cloudwatch.Dimension).SetName("host").SetValue("prod-host01"), new(cloudwatch.Dimension).SetName("vpc").SetValue("pub-us-east-1d")},
		},
		"more_than_10_labels": {
			model.Metric{model.MetricNameLabel: "foo", "label1": "1", "label2": "2", "label3": "3", "label4": "4", "label5": "5", "label6": "6", "label7": "7", "label8": "8", "label9": "9", "label10": "10", "label11": "11"},
			[]*cloudwatch.Dimension{
				new(cloudwatch.Dimension).SetName("label1").SetValue("1"),
				new(cloudwatch.Dimension).SetName("label10").SetValue("10"),
				new(cloudwatch.Dimension).SetName("label11").SetValue("11"),
				new(cloudwatch.Dimension).SetName("label2").SetValue("2"),
				new(cloudwatch.Dimension).SetName("label3").SetValue("3"),
				new(cloudwatch.Dimension).SetName("label4").SetValue("4"),
				new(cloudwatch.Dimension).SetName("label5").SetValue("5"),
				new(cloudwatch.Dimension).SetName("label6").SetValue("6"),
				new(cloudwatch.Dimension).SetName("label7").SetValue("7"),
				new(cloudwatch.Dimension).SetName("label8").SetValue("8"),
			},
		},
		"special_labels_ignored": {model.Metric{model.MetricNameLabel: "foo", cwUnitLabel: "Bytes", cwHighResLabel: ""}, []*cloudwatch.Dimension{}},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			actual := getDimensions(c.m)
			if !reflect.DeepEqual(actual, c.expected) {
				t.Errorf("got %+v; expected %+v", actual, c.expected)
			}
		})
	}
}

func Test_getResolution(t *testing.T) {
	cases := map[string]struct {
		m        model.Metric
		expected int64
	}{
		"default":  {model.Metric{}, 60},
		"high_res": {model.Metric{cwHighResLabel: ""}, 1},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			if actual := getResolution(c.m); actual != c.expected {
				t.Errorf("got %d; expected %d", actual, c.expected)
			}
		})
	}
}

func Test_getUnit(t *testing.T) {
	cases := map[string]struct {
		m        model.Metric
		expected string
	}{
		"default": {model.Metric{}, ""},
		"custom":  {model.Metric{cwUnitLabel: "Bytes"}, "Bytes"},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			if actual := getUnit(c.m); actual != c.expected {
				t.Errorf("got %q; expected %q", actual, c.expected)
			}
		})
	}
}
