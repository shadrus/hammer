package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTask(t *testing.T) {

	// assert equality
	configData := []byte(`
url: http://10.10.77.140:15500/create
method: GET
params:
  url: https://habr.com
headers:
  User-Agent: hammer
  Content-Type: application/json
timeout: 2
durability: 3
maxRPS: 5
body: '{"test": 3}'
growsCoef: 0.8`)
	task, err := NewTask(configData)
	assert.Equal(t, task.Timeout, 2, "they should be equal")
	assert.Nil(t, err)

	configData = []byte(`
{"url": "http://10.10.77.140:15500/create",
"method": "GET",
"params":{"url: https://habr.com"},
"headers": {"User-Agent": "hammer"},
"timeout": 2,
"durability": 3,
"maxRPS": 5,
"growsCoef": 0.8}`)
	task, err = NewTask(configData)
	assert.Equal(t, task.Timeout, 2, "they should be equal")
	assert.Nil(t, err)
	configData = []byte(`{"url": http://10.10.77.140:15500/create`)
	task, err = NewTask(configData)
	assert.NotNil(t, err)
}
