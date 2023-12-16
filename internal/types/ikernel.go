package types

type IKernel interface {
	DeleteGreensByPID(pid uint32)
	DeleteRodentByPID(pid uint32)
	IsReproductionPossible() (string, bool)
	NumberCorpseInc()
	NumberCorpseAdd(count int)
	NumberCorpseDec()
	GetNumberCorpse() int
}
