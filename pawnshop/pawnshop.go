package pawnshop

import "gorm.io/gorm"

type Pawnshop struct {
	gorm.Model
	NamaLengkap         string  // John Doe
	AsalBarangJaminan   *string // Hasil Usaha, Harta Pribadi, dll
	StatusTransaksi     *string // Untuk Diri Sendiri, Untuk Orang Lain, dll
	TujuanTransaksi     *string // Usaha. Investasi, dll
	CaraPembayaran      string  // Non Tunai
	FiturYangDiinginkan *string // Reguler, Bisnis, Fleksi
	PengambilanUang     string  // namaBank#nomorRekening#atasNama
	KelebihanLelang     *string
	BesarPinjaman       *string
	BarangJaminan       string
	Alamat              string
	NoHP                string
}

type MySqlDataStore struct {
	db *gorm.DB
}

func NewMySqlDataStore(db *gorm.DB) *MySqlDataStore {
	return &MySqlDataStore{db: db}
}

func (m *MySqlDataStore) AddPawnshop(pawnshop Pawnshop) error {
	result := m.db.Create(&pawnshop)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (m *MySqlDataStore) GetPawnshop() ([]Pawnshop, error) {
	var pawnshops []Pawnshop
	result := m.db.Find(&pawnshops)
	if result.Error != nil {
		return nil, result.Error
	}
	return pawnshops, nil
}

func (m *MySqlDataStore) GetPawnshopByID(id string) (*Pawnshop, error) {
	var pawnshop Pawnshop
	result := m.db.Where("id = ?", id).Find(&pawnshop)
	if result.Error != nil {
		return nil, result.Error
	}
	return &pawnshop, nil
}
