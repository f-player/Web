package repository

import (
	"fmt"
	"strings"
)

type Repository struct{}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Order struct {
	ID       int
	Title    string
	ImageURL string
	C        int
	N        int
}

// // Коллекция корзины
// var Cart []Order

// Коллекция услуг
func (r *Repository) GetOrders() ([]Order, error) {
	orders := []Order{
		{ID: 1, Title: "Наземные травоядные (C3)", ImageURL: "http://127.0.0.1:9000/imagegroup/i1.webp", C: -20, N: 5},
		{ID: 2, Title: "C3-злаки", ImageURL: "http://127.0.0.1:9000/imagegroup/i2.webp", C: -24, N: -5},
		{ID: 3, Title: "Морская рыба", ImageURL: "http://127.0.0.1:9000/imagegroup/i3.webp", C: -15, N: 5},
		{ID: 4, Title: "Пульсы / бобовые", ImageURL: "http://127.0.0.1:9000/imagegroup/i5.jpg", C: -30, N: 2},
		{ID: 5, Title: "Овощи", ImageURL: "http://127.0.0.1:9000/imagegroup/i4.webp", C: -28, N: 2},
	}

	if len(orders) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}
	return orders, nil
}

// Получение одной услуги
func (r *Repository) GetOrder(id int) (Order, error) {
	orders, _ := r.GetOrders()
	for _, o := range orders {
		if o.ID == id {
			return o, nil
		}
	}
	return Order{}, fmt.Errorf("заказ не найден")
}

// Поиск по названию
func (r *Repository) GetOrdersByTitle(title string) ([]Order, error) {
	orders, _ := r.GetOrders()
	var result []Order
	for _, o := range orders {
		if strings.Contains(strings.ToLower(o.Title), strings.ToLower(title)) {
			result = append(result, o)
		}
	}
	return result, nil
}

// // Работа с корзиной
// func (r *Repository) AddToCart(order Order) {
// 	Cart = append(Cart, order)
// }

// func (r *Repository) GetCart() []Order {
// 	return Cart
// }


func (r *Repository) GetCartOrder(id int) []Order {
  cartorder := map[int][]int{1: {1, 3, 5}}
  var orderInGroup []Order
  for _, orderID := range cartorder[id] {
    order, err := r.GetOrder(orderID)
    if err == nil {
      orderInGroup = append(orderInGroup, order)
    }
  }
  return orderInGroup
}

func (r *Repository) GetCartCount(id int) int {
  return len(r.GetCartOrder(id))
}

func (r *Repository) GetCartId() int {
  return 1
}
