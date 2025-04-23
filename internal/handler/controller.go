package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"swagger_petstore/internal/service"
	"swagger_petstore/middleware"
	"swagger_petstore/petstore"
	"swagger_petstore/responder"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/golang-jwt/jwt/v5"
)

type API struct {
	responder responder.Responder
	service   service.Servicer
}

func NewAPI(responder responder.Responder, service service.Servicer) *API {
	return &API{
		responder: responder,
		service:   service,
	}
}

// @Summary			create pet
// @Security 		ApiKeyAuth
// @Description		create pet
// @Tags			pet
// @Accept			json
// @Produce			json
// @Param			pet   body	petstore.Pet	true  "create pet"
// @Success			200		{object}	ResponseData
// @Router			/pet [post]
func (A *API) AddPet(w http.ResponseWriter, r *http.Request) {
	var pet petstore.Pet
	if err := json.NewDecoder(r.Body).Decode(&pet); err != nil {
		log.Fatal(err)
		return
	}

	err := A.service.AddPet(r.Context(), pet)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}
	A.responder.OutputJSON(w, ResponseData{
		Success: true,
		Data: Data{
			Message: fmt.Sprintf("added a new pet %s", pet.Name),
		},
	})
}

// @Summary			update pet
// @Security 		ApiKeyAuth
// @Description		update pet
// @Tags			pet
// @Accept			json
// @Produce			json
// @Param			pet   body	petstore.Pet	true  "update pet"
// @Success			200		{object}	ResponseData
// @Router			/pet [put]
func (A *API) UpdatePet(w http.ResponseWriter, r *http.Request) {
	var pet petstore.Pet
	if err := json.NewDecoder(r.Body).Decode(&pet); err != nil {
		log.Fatal(err)
		return
	}

	err := A.service.UpdatePet(r.Context(), pet)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}

	A.responder.OutputJSON(w, ResponseData{
		Success: true,
		Data: Data{
			Message: fmt.Sprintf("update pet %s", pet.Name),
		},
	})
}

// @Summary			find pets by status
// @Security 		ApiKeyAuth
// @Description		find pet
// @Tags			pet
// @Accept			json
// @Produce			json
// @Param			status   query	[]string	true  "Status values that need to be considered for filter" Enums(available,pending,sold)
// @Success			200		{object}	ResponsePets
// @Router			/pet/findByStatus [get]
func (A *API) FindPetsByStatus(w http.ResponseWriter, r *http.Request, params petstore.FindPetsByStatusParams) {
	pets, err := A.service.FindPetsByStatus(r.Context(), params)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}

	A.responder.OutputJSON(w, ResponsePets{
		Success: true,
		Data: PetsData{
			Pets: pets,
		},
	})
}

// @Summary			find pets by tags
// @Security		ApiKeyAuth
// @Description		find pet
// @Tags			pet
// @Accept			json
// @Produce			json
// @Param			tags   query	[]string	true  "find pet by tags"
// @Success			200		{object}	ResponsePets
// @Router			/pet/findByTags [get]
func (A *API) FindPetsByTags(w http.ResponseWriter, r *http.Request, params petstore.FindPetsByTagsParams) {
	pets, err := A.service.FindPetsByTags(r.Context(), params)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}

	A.responder.OutputJSON(w, ResponsePets{
		Success: true,
		Data: PetsData{
			Pets: pets,
		},
	})
}

// @Summary			delete pet
// @Security		ApiKeyAuth
// @Description		delete
// @Tags			pet
// @Accept			json
// @Produce			json
// @Param			petId   path	int	true  " "
// @Param			params   header	petstore.DeletePetParams	true  "delete pet"
// @Success			200		{object}	ResponseData
// @Router			/pet/{petId} [delete]
func (A *API) DeletePet(w http.ResponseWriter, r *http.Request, petId int64, params petstore.DeletePetParams) {
	if params.ApiKey == nil {
		return
	}

	err := A.service.DeletePet(r.Context(), petId, params)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}

	A.responder.OutputJSON(w, ResponseData{
		Success: true,
		Data: Data{
			Message: fmt.Sprintf("pet number %d has been deleted", petId),
		},
	})
}

// @Summary			find pet by id
// @Security		ApiKeyAuth
// @Description		get pet
// @Tags			pet
// @Accept			json
// @Produce			json
// @Param			petId   path	int	true  "pet id"
// @Success			200		{object}	ResponsePet
// @Router			/pet/{petId} [get]
func (A *API) GetPetById(w http.ResponseWriter, r *http.Request, petId int64) {
	pet, err := A.service.GetPetById(r.Context(), petId)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}

	A.responder.OutputJSON(w, ResponsePet{
		Success: true,
		Data: PetData{
			Message: fmt.Sprintf("found a pet with the id %d", petId),
			Pet:     pet,
		},
	})
}

// @Summary			update pet with form
// @Security 		ApiKeyAuth
// @Description		update
// @Tags			pet
// @Accept			x-www-form-urlencoded
// @Produce			json
// @Param			petId   path	    int		true   "ID of pet that needs to be updated"
// @Param			name    formData	string	false  "Updated name of the pet"
// @Param			status  formData	string	false  "Updated status of the pet"
// @Success			200		{object}	ResponseData
// @Router			/pet/{petId} [post]
func (A *API) UpdatePetWithForm(w http.ResponseWriter, r *http.Request, petId int64, params petstore.UpdatePetWithFormParams) {
	name := r.FormValue("name")
	status := r.FormValue("status")
	params.Name = &name
	params.Status = &status

	err := A.service.UpdatePetWithForm(r.Context(), petId, params)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}

	A.responder.OutputJSON(w, ResponseData{
		Success: true,
		Data: Data{
			Message: fmt.Sprintf("pet number %d has been update", petId),
		},
	})
}

func (A *API) UploadFile(w http.ResponseWriter, r *http.Request, petId int64, params petstore.UploadFileParams) {
	//TODO implement me
	panic("implement me")
}

// @Summary			get inventory
// @Security 		ApiKeyAuth
// @Description		get
// @Tags			store
// @Accept			json
// @Produce			json
// @Success			200		{object}	ResponseInventory
// @Router			/store/inventory [get]
func (A *API) GetInventory(w http.ResponseWriter, r *http.Request) {
	inventory, err := A.service.GetInventory(r.Context())
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}

	A.responder.OutputJSON(w, ResponseInventory{
		Success: true,
		Data: InventoryData{
			Message:   fmt.Sprint("received a mapping of status codes from received"),
			Inventory: inventory,
		},
	})
}

// @Summary			place order
// @Description		create order
// @Tags			store
// @Accept			json
// @Produce			json
// @Param			order   body	petstore.Order	true  "order"
// @Success			200		{object}	ResponseOrder
// @Router			/store/order [post]
func (A *API) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var order petstore.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		A.responder.ErrorBadRequest(w, err)
		return
	}
	err := A.service.PlaceOrder(r.Context(), order)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}
	A.responder.OutputJSON(w, ResponseOrder{
		Success: true,
		Data: OrderData{
			Message: fmt.Sprint("the order has been created"),
			Order:   order,
		},
	})
}

// @Summary			delete order
// @Description		delete
// @Tags			store
// @Accept			json
// @Produce			json
// @Param			orderId   path	int	true  "id"
// @Success			200		{object}	ResponseData
// @Router			/store/order/{orderId} [delete]
func (A *API) DeleteOrder(w http.ResponseWriter, r *http.Request, orderId int64) {
	err := A.service.DeleteOrder(r.Context(), orderId)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}

	A.responder.OutputJSON(w, ResponseData{
		Success: true,
		Data: Data{
			Message: fmt.Sprintf("order number %d has been delete", orderId),
		},
	})
}

// @Summary			get order
// @Description		get
// @Tags			store
// @Accept			json
// @Produce			json
// @Param			orderId   path	int	true  "id"
// @Success			200		{object}	ResponseOrder
// @Router			/store/order/{orderId} [get]
func (A *API) GetOrderById(w http.ResponseWriter, r *http.Request, orderId int64) {
	order, err := A.service.GetOrderById(r.Context(), orderId)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}

	A.responder.OutputJSON(w, ResponseOrder{
		Success: true,
		Data: OrderData{
			Order: order,
		},
	})
}

// @Summary			create user
// @Description		created user object
// @Tags			user
// @Accept			json
// @Produce			json
// @Param			user	body	petstore.User	true	"create user"
// @Success			200		{object}	ResponseUser
// @Router			/user [post]
func (A *API) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user petstore.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		A.responder.ErrorBadRequest(w, err)
		return
	}

	err := A.service.CreateUser(r.Context(), user)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}

	A.responder.OutputJSON(w, ResponseUser{
		Success: true,
		Data: UserData{
			User: user,
		},
	})
}

// @Summary			create list user
// @Description		create
// @Tags			user
// @Accept			json
// @Produce			json
// @Param			user	body	[]petstore.User	true	"create list user"
// @Success			200		{object}	ResponseUsers
// @Router			/user/createWithList [post]
func (A *API) CreateUsersWithListInput(w http.ResponseWriter, r *http.Request) {
	var users []petstore.User
	if err := json.NewDecoder(r.Body).Decode(&users); err != nil {
		A.responder.ErrorBadRequest(w, err)
		return
	}

	err := A.service.CreateUsersWithListInput(r.Context(), users)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}

	A.responder.OutputJSON(w, ResponseUsers{
		Success: true,
		Data: UsersData{
			Users: users,
		},
	})
}

// @Summary			login user
// @Description		auth
// @Tags			user
// @Accept			json
// @Produce			json
// @Param			user	query	petstore.LoginUserParams	true	"create list user"
// @Success			200		{object}	AuthResponse
// @Router			/user/login [get]
func (A *API) LoginUser(w http.ResponseWriter, r *http.Request, params petstore.LoginUserParams) {
	token, err := A.service.LoginUser(r.Context(), params)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}

	A.responder.OutputJSON(w, AuthResponse{
		Success: true,
		Data: LoginData{
			Message: "Bearer token",
			Token:   token,
		},
	})
}

// @Summary			logout user
// @Security 		ApiKeyAuth
// @Description		logout
// @Tags			user
// @Accept			json
// @Produce			json
// @Success			200		{object}	ResponseData
// @Router			/user/logout [get]
func (A *API) LogoutUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tokenString := jwtauth.TokenFromHeader(r)

	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		fmt.Printf("Ошибка парсинга токена: %v\n", err)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		tokenID, ok := claims["jti"].(string)
		if !ok || tokenID == "" {
			http.Error(w, "token missing ID", http.StatusBadRequest)
			return
		}

		exp, ok := claims["exp"].(float64)
		if !ok {
			http.Error(w, "invalid token expiration", http.StatusBadRequest)
			return
		}
		expTime := time.Unix(int64(exp), 0)
		err = A.service.LogoutUser(ctx, tokenID, tokenString, expTime)
		if err != nil {
			A.responder.ErrorInternal(w, err)
			return
		}

		middleware.ClearTokenFromClient(w)

		A.responder.OutputJSON(w, ResponseData{
			Success: true,
			Data: Data{
				Message: fmt.Sprint("successfully logged out"),
			},
		})
	} else {
		fmt.Println("Не удалось извлечь claims из токена")
	}
}

// @Summary			delete user
// @Description		delete
// @Tags			user
// @Accept			json
// @Produce			json
// @Param			username   path	string	true  "name"
// @Success			200		{object}	ResponseData
// @Router			/user/{username} [delete]
func (A *API) DeleteUser(w http.ResponseWriter, r *http.Request, username string) {
	err := A.service.DeleteUser(r.Context(), username)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}

	A.responder.OutputJSON(w, ResponseData{
		Success: true,
		Data: Data{
			Message: fmt.Sprintf("deleted user %s", username),
		},
	})
}

// @Summary			get user by name
// @Description		get user
// @Tags			user
// @Accept			json
// @Produce			json
// @Param			username	path	string	true "get user"
// @Success			200		{object}	ResponseUser
// @Router			/user/{username} [get]
func (A *API) GetUserByName(w http.ResponseWriter, r *http.Request, username string) {
	user, err := A.service.GetUserByName(r.Context(), username)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}

	A.responder.OutputJSON(w, ResponseUser{
		Success: true,
		Data: UserData{
			User: user,
		},
	})
}

// @Summary			update user
// @Description		update
// @Tags			user
// @Accept			json
// @Produce			json
// @Param			username	path string	true "username"
// @Param			user		body petstore.User	true "user"
// @Success			200			{object}	ResponseData
// @Router			/user/{username} [put]
func (A *API) UpdateUser(w http.ResponseWriter, r *http.Request, username string) {
	var user petstore.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		A.responder.ErrorBadRequest(w, err)
		return
	}
	fmt.Println(username)
	_, err := A.service.GetUserByName(r.Context(), username)
	if err != nil {
		A.responder.OutputJSON(w, ErrorResponse{
			Success:   false,
			ErrorCode: http.StatusNotFound,
			Data: Data{
				Message: fmt.Sprint("user not found"),
			},
		})
		return
	}

	err = A.service.UpdateUser(r.Context(), user)
	if err != nil {
		A.responder.ErrorInternal(w, err)
		return
	}

	A.responder.OutputJSON(w, ResponseData{
		Success: true,
		Data: Data{
			Message: fmt.Sprintf("user %s has been updated", username),
		},
	})
}
