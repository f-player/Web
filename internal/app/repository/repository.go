package repository

import (

   "gorm.io/driver/postgres"
   "gorm.io/gorm"
)

type Repository struct {
   db *gorm.DB
}

func New(dsn string) (*Repository, error) {
   db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // подключаемся к БД
   if err != nil {
      return nil, err
   }

   // Возвращаем объект Repository с подключенной базой данных
   return &Repository{
      db: db,
   }, nil
}

// type Product struct {
// 	ID       int
// 	Title    string
// 	ImageURL string
// 	C        int
// 	N        int
// }

// // Коллекция корзины
// var Cart []Order

// Коллекция услуг
// func (r *Repository) GetProducts() ([]Product, error) {
// 	products := []Product{
// 		{ID: 1, Title: "Наземные травоядные (C3)", ImageURL: "http://127.0.0.1:9000/imagegroup/i1.webp", C: -20, N: 5},
// 		{ID: 2, Title: "C3-злаки", ImageURL: "http://127.0.0.1:9000/imagegroup/i2.webp", C: -24, N: -5},
// 		{ID: 3, Title: "Морская рыба", ImageURL: "http://127.0.0.1:9000/imagegroup/i3.webp", C: -15, N: 5},
// 		{ID: 4, Title: "Пульсы / бобовые", ImageURL: "http://127.0.0.1:9000/imagegroup/i5.jpg", C: -30, N: 2},
// 		{ID: 5, Title: "Овощи", ImageURL: "http://127.0.0.1:9000/imagegroup/i4.webp", C: -28, N: 2},
// 	}

// 	if len(products) == 0 {
// 		return nil, fmt.Errorf("массив пустой")
// 	}
// 	return products, nil
// }



// // Работа с корзиной
// func (r *Repository) AddToCart(order Order) {
// 	Cart = append(Cart, order)
// }

// func (r *Repository) GetCart() []Order {
// 	return Cart
// }


// func (r *Repository) GetCalculationProduct(id int) []Product {
//   calculationproduct := map[int][]int{1: {1, 3, 5}}
//   var productInGroup []Product
//   for _, productID := range calculationproduct[id] {
//     product, err := r.GetProduct(productID)
//     if err == nil {
//       productInGroup = append(productInGroup, product)
//     }
//   }
//   return productInGroup
// }

// func (r *Repository) GetCalculationCount(id int) int {
//   return len(r.GetCalculationProduct(id))
// }

// func (r *Repository) GetCalculationId() int {
//   return 1
// }
