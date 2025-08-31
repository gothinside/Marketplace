package test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hw11_shopql/graph"
	"hw11_shopql/graph/model"
	"hw11_shopql/pkg/cart"
	"hw11_shopql/pkg/catalog"
	"hw11_shopql/pkg/comment"
	"hw11_shopql/pkg/item"
	"hw11_shopql/pkg/order"
	"hw11_shopql/pkg/rate"
	"hw11_shopql/pkg/role"
	"hw11_shopql/pkg/seller"
	"hw11_shopql/pkg/session"
	"hw11_shopql/pkg/user"
	"hw11_shopql/pkg/utils/roleutils"
	"hw11_shopql/pkg/utils/sessionutils"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	defaultPort   = "8080"
	mongoURI      = "mongodb://127.0.0.1:27017"
	mongoAuthDB   = "admin"
	mongoUsername = "root"
	mongoPassword = "example"
	databaseName  = "hz"
)

type Resp map[string]map[string]string

type Catalog struct {
	ID     int         `json:"id"`
	Name   string      `json:"name"`
	Childs []*Category `json:"childs,omitempty"`
	Items  []*Item     `json:"items,omitempty"`
}

type Category struct {
	ID     int         `json:"id"`
	Name   string      `json:"name"`
	Childs []*Category `json:"childs,omitempty"`
	Items  []*Item     `json:"items,omitempty"`
}

type Item struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	InStock  int    `json:"in_stock"`
	SellerID int    `json:"seller_id"`
}

type Data struct {
	Catalog model.Catalog  `json:"catalog"`
	Sellers []model.Seller `json:"sellers"`
}

type Seller struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Deals int    `json:"deals"`
}

// type UserStorage interface {
// 	AddUser(user *user.User) error
// }

// type UserHandler struct {
// 	St UserStorage
// }

// func (uh *UserHandler) registration(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == "POST" {
// 		user := &user.User{}
// 		err := json.NewDecoder(r.Body).Decode(&user)
// 		if err != nil {
// 			fmt.Println(err)
// 		}
// 		uh.St.AddUser(user)
// 		res, _ := json.Marshal(Resp{"body": map[string]string{"token": "1"}})
// 		w.Write(res)
// 		w.Header().Add("body.token", "1")
// 		w.WriteHeader(200)
// 	} else {
// 		w.WriteHeader(401)
// 	}
// }

func Middleware(sm session.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := sm.Check(r)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), "tokens", session)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func connectMongoDB() (*mongo.Client, error) {
	credential := options.Credential{
		AuthSource: mongoAuthDB,
		Username:   mongoUsername,
		Password:   mongoPassword,
	}

	clientOpts := options.Client().
		ApplyURI(mongoURI).
		SetAuth(credential)

	client, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}

func loadTestData(filename string) (*Data, error) {
	data := &Data{}
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read test data file: %w", err)
	}

	if err := json.Unmarshal(fileContent, data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal test data: %w", err)
	}

	return data, nil
}

func GetApp() *chi.Mux {
	client, err := connectMongoDB()
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}
	testData, err := loadTestData("testdata.json")
	if err != nil {
		log.Printf("Warning: failed to load test data: %v", err)
	} else {
		fmt.Printf("Loaded test data: %+v\n", testData)
	}
	// Initialize catalog handler
	db := client.Database(databaseName)
	collection := db.Collection("Catalogs")
	item_collection := db.Collection("Items")
	rateCollection := db.Collection("Rates")
	rateRepos := *rate.CreateRateRepo(rateCollection)
	commentCollection := db.Collection("Comments")
	commRepo := *comment.CreateCommentRepo(commentCollection)
	itemHandler := item.CreateItemsHandler(item_collection, &rateRepos, &commRepo)
	catalogHandler := catalog.CreateCatalogHandler(collection, itemHandler)
	cartCollection := db.Collection("Carts")
	cartRepos := *cart.CreateCartRepo(cartCollection, itemHandler)
	seller_collection := db.Collection("Sellers")
	sellerHandler := seller.CreateSellersHandler(seller_collection)
	orderCollection := db.Collection("orders")
	orderRepo := *order.CreateOrderRepo(orderCollection, &cartRepos, itemHandler)
	// Insert test data if available
	if testData != nil {
		if err := catalogHandler.AddNewCatalog(context.Background(), testData.Catalog); err != nil {
			log.Printf("Failed to insert catalog: %v", err)
		}
		if err := itemHandler.InsertCatalogsItems(context.Background(), testData.Catalog); err != nil {
			log.Printf("Failed to insert catalog: %v", err)
		}
		for _, seller := range testData.Sellers {
			limit := 5
			offset := 0
			items, err := itemHandler.GetItemsBySellerID(context.Background(), seller.ID, &limit, &offset)
			if err != nil {
				for _, item := range items {
					seller.ItemIds = append(seller.ItemIds, item.ID)
				}
			}
			if err != nil {
				log.Printf("Failed to insert seller: %v", err)
			}
		}
		// Example lookup
		catalog, err := catalogHandler.LookupCatalog(context.Background(), 1)
		if err != nil {
			log.Printf("Failed to lookup catalog: %v", err)
		} else {
			fmt.Printf("Found catalog: %+v\n", catalog)
		}
	}
	psqlInfo := fmt.Sprintf("user=%s "+
		"password=%s dbname=%s sslmode=disable",
		username, password, dbname)
	postgre, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	c := graph.Config{Resolvers: &graph.Resolver{CatalogRepo: catalogHandler,
		CartRepo:   &cartRepos,
		ItemRepo:   itemHandler,
		SellerRepo: sellerHandler,
		OrderRepo:  &orderRepo,
	}}
	c.Directives.Authorized = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		if ctx.Value("tokens") == nil {
			graphql.AddError(ctx, fmt.Errorf("User not authorized"))
			return err, nil
		}
		return next(ctx)
	}
	c.Directives.HasRole = func(ctx context.Context, obj interface{}, next graphql.Resolver, role model.Role) (res interface{}, err error) {
		if id, err := sessionutils.IdFromContex(ctx); err == nil {
			ok := roleutils.HasRole(postgre, id, role.String())
			if !ok {
				graphql.AddError(ctx, fmt.Errorf("Forbiden"))
				return err, nil
			}
		} else {
			return nil, nil
		}
		return next(ctx)
	}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(c))
	sm := session.NewSessionsDB(postgre)
	router := chi.NewRouter()
	router.Use(Middleware(sm))
	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)
	roleRepo := role.CreateRoleRepo(postgre)
	ur := user.CreateUserRepo(postgre, roleRepo)
	uh := user.CreateUserHandler(ur, sm)
	router.HandleFunc("/register", uh.Reg)
	router.HandleFunc("/login", uh.Log)
	log.Printf("Connect to http://localhost:%v/ for GraphQL playground", port)
	return router
}

const (
	host     = "localhost"
	port     = 5432
	username = "postgres"
	password = "123"
	dbname   = "shopql"
)
