package keypair

// 存储器方法合集
type Storer interface {
	Open(dataPath string)
	Append(PublicKey) error
	Close() error
}

// 组合存储器
type CombineStorer struct {
	storers []Storer
}

func NewCombineStorer(storers ...Storer) *CombineStorer {
	return &CombineStorer{
		storers: storers,
	}
}

func (s *CombineStorer) Open(dataPath string) {
	for _, st := range s.storers {
		st.Open(dataPath)
	}
}

func (s *CombineStorer) Append(pubKey PublicKey) error {
	for _, st := range s.storers {
		err := st.Append(pubKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *CombineStorer) Close() error {
	for _, st := range s.storers {
		err := st.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
