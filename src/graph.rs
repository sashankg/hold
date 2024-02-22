use core::hash;
use std::collections::{BinaryHeap, HashSet};

use anyhow::anyhow;
use graphql_parser::schema::{
    self, parse_schema, Definition, Document, ObjectType, Type, TypeDefinition,
};

enum MigrationStep<'a> {
    CreateTable(&'a GQLObjectType<'a>),
    DeleteTable(&'a str),
    AddColumn(&'a GQLField<'a>),
    ModifyColumn(&'a GQLField<'a>),
    RemoveColumn,
}

pub fn migrate(schema: &str) -> Result<(), anyhow::Error> {
    let ast = parse_schema::<&str>(schema)?;

    Ok(())
}

#[derive(Hash, Eq, PartialEq)]
enum GQLType<'a> {
    Int,
    Boolean,
    String,
    Float,
    ID,
    Named(&'a str),
}

#[derive(Hash, Eq)]
struct GQLField<'a> {
    name: &'a str,
    ty: GQLType<'a>,
    required: bool,
    array: bool,
}

impl PartialEq for GQLField<'_> {
    fn eq(&self, other: &Self) -> bool {
        self.name == other.name
    }
}

#[derive(Eq)]
struct GQLObjectType<'a> {
    name: &'a str,
    fields: HashSet<GQLField<'a>>,
}

impl PartialEq for GQLObjectType<'_> {
    fn eq(&self, other: &Self) -> bool {
        self.name == other.name
    }
}

impl hash::Hash for GQLObjectType<'_> {
    fn hash<H: hash::Hasher>(&self, state: &mut H) {
        self.name.hash(state);
    }
}

fn schema_type_to_gql_type<'a>(ty: &Type<'a, &'a str>) -> (GQLType<'a>, bool, bool) {
    match ty {
        schema::Type::NamedType(name) => (
            match *name {
                "Int" => GQLType::Int,
                "Boolean" => GQLType::Boolean,
                "String" => GQLType::String,
                "Float" => GQLType::Float,
                "ID" => GQLType::ID,
                _ => GQLType::Named(name),
            },
            false,
            false,
        ),
        schema::Type::NonNullType(ty) => {
            let (ty, _, array) = schema_type_to_gql_type(ty);
            (ty, true, array)
        }
        schema::Type::ListType(ty) => {
            let (ty, required, _) = schema_type_to_gql_type(ty);
            (ty, required, true)
        }
    }
}

fn schema_doc_to_gql_objects<'a>(ast: &Document<'a, &'a str>) -> HashSet<GQLObjectType<'a>> {
    let mut ir = HashSet::new();
    for def in ast.definitions.iter() {
        if let Definition::TypeDefinition(TypeDefinition::Object(ObjectType {
            name, fields, ..
        })) = def
        {
            let mut ir_fields = HashSet::new();
            for field in fields.iter() {
                let (ty, required, array) = schema_type_to_gql_type(&field.field_type);
                ir_fields.insert(GQLField {
                    name: field.name,
                    ty,
                    required,
                    array,
                });
            }
            ir.insert(GQLObjectType {
                name,
                fields: ir_fields,
            });
        }
    }
    ir
}

fn diff<'a>(
    original: &'a HashSet<GQLObjectType<'a>>,
    new: &'a HashSet<GQLObjectType<'a>>,
) -> Vec<MigrationStep<'a>> {
    let mut steps = Vec::new();
    new.difference(original)
        .for_each(|obj| steps.push(MigrationStep::CreateTable(obj)));
    original.difference(new).for_each(|obj| {
        steps.push(MigrationStep::DeleteTable(obj.name));
    });
    new.intersection(original).for_each(|new_obj| {
        let original_obj = original.get(new_obj).unwrap();
        new_obj
            .fields
            .difference(&original_obj.fields)
            .for_each(|field| {
                steps.push(MigrationStep::AddColumn(field));
            });
        original_obj
            .fields
            .difference(&new_obj.fields)
            .for_each(|_| {
                steps.push(MigrationStep::RemoveColumn);
            });
    });
    steps
}

#[test]
fn test_migrate() -> Result<(), anyhow::Error> {
    migrate(
        r#"
        type User {
            id: ID!
            name: Name
        }
    "#,
    )
}
