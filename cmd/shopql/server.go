package main

import (
	"context"
	"database/sql"
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

type Data struct {
	Catalog model.Catalog  `json:"catalog"`
	Sellers []model.Seller `json:"sellers"`
}

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

const (
	host     = "localhost"
	port     = 5432
	username = "postgres"
	password = "123"
	dbname   = "shopql"
)

func main() {
	client, err := connectMongoDB()
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
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
	psqlInfo := fmt.Sprintf("user=%s "+
		"password=%s dbname=%s sslmode=disable",
		username, password, dbname)
	postgre, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer postgre.Close()

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
	log.Fatal(http.ListenAndServe(":8080", router))
}
