package repository

import "hot-coffee/models"

type OrderRepository interface {
	Create(order models.Order) error
	FindAll() ([]models.Order, error)
	FindById(id string) (*models.Order, error)
	Update(id string) error
	Delete(id string) error
	Close(id string) error
}

type jsonOrderRepository struct {
	dataDir string
}

func (j jsonOrderRepository) Create(order models.Order) error {
	//TODO implement me
	panic("implement me")
}

func (j jsonOrderRepository) FindAll() ([]models.Order, error) {
	//TODO implement me
	panic("implement me")
}

func (j jsonOrderRepository) FindById(id string) (*models.Order, error) {
	//TODO implement me
	panic("implement me")
}

func (j jsonOrderRepository) Update(id string) error {
	//TODO implement me
	panic("implement me")
}

func (j jsonOrderRepository) Delete(id string) error {
	//TODO implement me
	panic("implement me")
}

func (j jsonOrderRepository) Close(id string) error {
	//TODO implement me
	panic("implement me")
}

func NewJsonOrderRepository(dir string) OrderRepository {
	return &jsonOrderRepository{dataDir: dir}
}
