package goseq

type Region byte

const (
	USEast       Region = 0x00
	USWest              = 0x01
	SouthAmerica        = 0x02
	Europe              = 0x03
	Asia                = 0x04
	Australia           = 0x05
	MiddleEast          = 0x06
	Africa              = 0x07
	RestOfWorld         = 0xFF
)

type MasterServer interface {
}
