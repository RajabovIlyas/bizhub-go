package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/controllers/v1"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupV1AuthRoutes(router fiber.Router) {

	auth := router.Group("/auth")
	auth.Get("/favorites",
		middlewares.DeSerializeCustomer,
		controllers.GetCustomerFavorites)
	auth.Post("/favorite",
		middlewares.DeSerializeCustomer,
		controllers.AddFavorite)
	auth.Delete("/favorite",
		middlewares.DeSerializeCustomer,
		controllers.DeleteFavorite)
	// auth.Post("/save_favorites", middlewares.DeSerializeCustomer, controllers.SaveFavorites)
	auth.Post("/customer_login",
		controllers.AuthCustomerLogin)
	// auth.Get("/me",
	// 	middlewares.DeSerializeCustomer,
	// 	controllers.GetMe)
	// auth.Get("/ip",
	// 	controllers.GetMyIp)
	auth.Post("/refresh",
		controllers.RefreshAccessToken)
	auth.Put("/customer_profile",
		middlewares.DeSerializeCustomer,
		controllers.UpdateCustomerProfile)
	auth.Put("/change_password",
		middlewares.DeSerializeCustomer,
		controllers.ChangePassword)
	auth.Post("/recover_password",
		controllers.RecoverPassword)
	auth.Post("/has_phone",
		controllers.HasPhone)
	auth.Post("/become_seller",
		middlewares.DeSerializeCustomer,
		controllers.BecomeSeller)
	auth.Post("/signup",
		controllers.Signup)
	auth.Delete("/profile",
		middlewares.DeSerializeCustomer,
		controllers.DeleteProfile)
	auth.Post("/validate_password",
		middlewares.DeSerializeCustomer,
		controllers.ValidatePassword)
	auth.Post("/logout",
		middlewares.DeSerializeCustomer,
		controllers.Logout)

	seller_profile := auth.Group("seller_profile",
		middlewares.DeSerializeCustomer,
		middlewares.AllowSeller())
	seller_profile.Get("/",
		controllers.GetSellerProfile)
	seller_profile.Put("/",
		controllers.UpdateSellerProfile)
	seller_profile.Get("/products",
		controllers.GetSellerProfileProducts)
	seller_profile.Get("/products/:id",
		controllers.GetProductForEditing)
	seller_profile.Put("/products/:id",
		controllers.EditProduct)
	seller_profile.Delete("/products/:id",
		controllers.DeleteProduct)
	seller_profile.Delete("/posts/:id",
		controllers.DeletePost)
	seller_profile.Post("/products",
		controllers.AddNewProduct)
	seller_profile.Get("/posts",
		controllers.GetSellerProfilePosts)
	seller_profile.Post("/posts",
		controllers.AddNewPost)
	seller_profile.Get("/categories",
		controllers.GetSellerProfileCategories)
	seller_profile.Get("/products_post",
		controllers.GetRelatedProductsForPost)

}
