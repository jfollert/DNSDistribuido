package main

import (
	"testing"
	"context"
	pb "github.com/jfomu/DNSDistribuido/internal/proto"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	s := new(Server)

	consulta := new(pb.Consulta)
	consulta.NombreDominio = "google.com"
	consulta.Ip = "8.8.8.8"
	respuesta, _ := s.Create(context.Background(), consulta)




	assert.Equal(t, respuesta.Reloj, []int32{0,0,0})
}
