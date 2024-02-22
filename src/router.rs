mod register;

use std::sync::{Arc, Mutex};

use anyhow::anyhow;
use axum::{extract::State, routing::*, Extension};
use graphql_parser::query::Field;
use juniper::{graphql_value, Context, ExecutionResult, Executor, FieldError, GraphQLObject};
use juniper_axum::{extract::JuniperRequest, response::JuniperResponse};
use rusqlite::Connection;

pub struct AppState {
    conn: Mutex<Connection>,
}

impl Context for AppState {}

pub fn new_router(conn: Connection) -> axum::Router {
    let state = AppState {
        conn: Mutex::new(conn),
    };
    let schema = Schema::new(Query, Mutation::new(), juniper::EmptySubscription::new());
    Router::new()
        .route("/", get(|| async { "Hello, World!" }))
        .route(
            "/graphql",
            on(MethodFilter::GET.or(MethodFilter::POST), graphql_handler),
        )
        .route(
            "/graphiql",
            get(juniper_axum::graphiql("/graphql", "/subscriptions")),
        )
        .route(
            "/playground",
            get(juniper_axum::playground("graphql", "/subscriptions")),
        )
        .layer(Extension(Arc::new(schema)))
        .with_state(Arc::new(state))
}

async fn graphql_handler(
    Extension(schema): Extension<Arc<Schema>>,
    State(state): State<Arc<AppState>>,
    JuniperRequest(request): JuniperRequest,
) -> JuniperResponse {
    println!("request: {:?}", request);
    JuniperResponse(request.execute(&schema, &state).await)
}

type Schema = juniper::RootNode<'static, Query, Mutation, juniper::EmptySubscription<AppState>>;

#[derive(GraphQLObject)]
struct Kv {
    value: String,
}

#[derive(Clone, Copy, Debug)]
pub struct Query;

#[juniper::graphql_object(context = AppState)]
impl Query {
    /// Adds two `a` and `b` numbers.
    fn add(a: i32, b: i32) -> i32 {
        a + b
    }

    fn kv(context: &AppState, key: String) -> Kv {
        let x = context
            .conn
            .lock()
            .unwrap()
            .query_row("SELECT json(value) from kv where key = $1", [key], |row| {
                row.get::<usize, String>(0)
            })
            .unwrap();
        Kv { value: x }
    }
}

#[derive(Clone, Copy, Debug, GraphQLObject)]
#[graphql(context = AppState)]
pub struct Mutation {
    kv: KvMutation,
}

impl Mutation {
    fn new() -> Self {
        Mutation { kv: KvMutation }
    }
}

#[derive(Clone, Copy, Debug)]
pub struct KvMutation;

#[juniper::graphql_object(context = AppState)]
impl KvMutation {
    fn put<S: juniper::ScalarValue>(
        context: &AppState,
        key: String,
        value: String,
        executor: &Executor<'_, '_, AppState, S>,
    ) -> Option<Kv> {
        match context.conn.lock().unwrap().execute(
            "INSERT into kv (key, value) VALUES ($1, jsonb($2))",
            [key, value.clone()],
        ) {
            Ok(_) => Some(Kv { value }),
            Err(e) => {
                executor.push_error(FieldError::new(&e, graphql_value!({ "internal": "error"})));
                None
            }
        }
    }
}
