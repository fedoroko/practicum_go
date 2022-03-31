package storage

import (
	"errors"
	"fmt"
)

var dummyGaugeStorage gaugeStorage

var dummyCounterStorage counterStorage

func Store(t string, n string, v string) error {
	if t == "gauge" {
		return update(&dummyGaugeStorage, n, v)
	} else if t == "counter" {
		return update(&dummyCounterStorage, n, v)
	}

	return errors.New("wrong type: " + t)
}

func Get(t string, n string) (string, error) {
	if t == "gauge" {
		return collect(&dummyGaugeStorage, n)
	} else if t == "counter" {
		return collect(&dummyCounterStorage, n)
	}

	return "", errors.New("wrong type")
}

func Values() (string, error) {
	ret := "Known values: \n"
	for n := range dummyGaugeStorage.Fields {
		v, err := dummyGaugeStorage.get(n)
		if err != nil {
			return ret, err
		}
		ret += fmt.Sprintf("%s - %s\n", n, v)
	}
	for n := range dummyCounterStorage.Fields {
		v, err := dummyCounterStorage.get(n)
		if err != nil {
			return ret, err
		}
		ret += fmt.Sprintf("%s - %s\n", n, v)
	}

	return ret, nil
}
