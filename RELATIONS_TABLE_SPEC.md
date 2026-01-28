# Relations Table Specification
## Naming Convention & Semantic Meaning

---

## 📋 Table Schema

```sql
CREATE TABLE relations (
  id                  BIGINT PRIMARY KEY,
  uuid                CHAR(36) UNIQUE,
  source_entity_id    BIGINT NOT NULL,      -- Entity yang "memiliki" relasi
  source_field_name   VARCHAR(100) NOT NULL, -- FK field name (di entity yang punya FK)
  target_entity_id    BIGINT NOT NULL,      -- Entity yang "dirujuk"
  relation_type       VARCHAR(50) NOT NULL, -- belongsTo, hasOne, hasMany, manyToMany
  on_delete           VARCHAR(50),          -- CASCADE, RESTRICT, SET NULL
  on_update           VARCHAR(50),          -- CASCADE, RESTRICT, SET NULL
  junction_table      VARCHAR(100),         -- For manyToMany only
  description         TEXT,
  required            TINYINT(1),
  created_at          DATETIME,
  updated_at          DATETIME,
  deleted_at          DATETIME
);
```

---

## 🎯 Naming Convention Rules

### **General Principle:**
- **source_entity_id** = Entity yang **"melakukan"** action (has, belongs to)
- **target_entity_id** = Entity yang **"dikenai"** action (is being referred to)
- **source_field_name** = FK field name di entity yang **memiliki FK column**

### **Key Point:**
⚠️ **`source_field_name` TIDAK selalu ada di `source_entity`!**
- Untuk `belongsTo`: FK ada di source entity
- Untuk `hasOne/hasMany`: FK ada di **target entity** (bukan source!)
- Field name tetap disimpan di `source_field_name` untuk konsistensi API

---

## 📚 Detailed Examples

### **Example 1: belongsTo**

**Scenario:** Post belongs to User

```
User (users)         Post (posts)
  id                   id
  name                 title
                       user_id  ← FK column
```

**Relations Table:**
```sql
source_entity_id    = Post.id (yang punya FK)
target_entity_id    = User.id (yang dirujuk)
source_field_name   = "user_id"
relation_type       = "belongsTo"
```

**Logic:**
- Post (source) **belongs to** User (target)
- FK `user_id` ada di table **posts** (source entity)
- `source_field_name` menunjuk ke field di **posts**

**Migration Generated:**
```sql
CREATE TABLE posts (
  ...
  user_id BIGINT NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id)
);
```

---

### **Example 2: hasMany**

**Scenario:** Product has many Stock

```
Product (products)   Stock (stocks)
  id                   id
  name                 quantity
                       product_id  ← FK column
```

**Relations Table:**
```sql
source_entity_id    = Product.id (yang "punya" banyak)
target_entity_id    = Stock.id (yang "dimiliki")
source_field_name   = "product_id"
relation_type       = "hasMany"
```

**Logic:**
- Product (source) **has many** Stock (target)
- FK `product_id` ada di table **stocks** (target entity) ⚠️
- `source_field_name` tetap `"product_id"` (nama field di stocks)

**Migration Generated:**
```sql
CREATE TABLE stocks (
  ...
  product_id BIGINT NOT NULL,
  FOREIGN KEY (product_id) REFERENCES products(id)
);
```

**⚠️ Important:**
- Meskipun field ada di **target entity (stocks)**
- Nama field tetap disimpan di **`source_field_name`**
- Ini untuk API consistency: "Product hasMany Stock via product_id"

---

### **Example 3: hasOne**

**Scenario:** User has one Profile

```
User (users)         Profile (profiles)
  id                   id
  name                 bio
                       user_id  ← FK column
```

**Relations Table:**
```sql
source_entity_id    = User.id (yang "punya")
target_entity_id    = Profile.id (yang "dimiliki")
source_field_name   = "user_id"
relation_type       = "hasOne"
```

**Logic:**
- User (source) **has one** Profile (target)
- FK `user_id` ada di table **profiles** (target entity) ⚠️
- `source_field_name` tetap `"user_id"`

**Migration Generated:**
```sql
CREATE TABLE profiles (
  ...
  user_id BIGINT NOT NULL UNIQUE,  -- UNIQUE untuk hasOne
  FOREIGN KEY (user_id) REFERENCES users(id)
);
```

---

### **Example 4: manyToMany**

**Scenario:** Post has many Tags (bidirectional)

```
Post (posts)         PostTag (post_tags)         Tag (tags)
  id                   post_id  ← FK               id
  title                tag_id   ← FK               name
```

**Relations Table:**
```sql
source_entity_id    = Post.id
target_entity_id    = Tag.id
source_field_name   = NULL (not used for manyToMany)
relation_type       = "manyToMany"
junction_table      = "post_tags"
```

**Logic:**
- Post (source) **has many** Tags (target) through junction
- No direct FK in either table
- Junction table `post_tags` holds both FKs

**Migration Generated:**
```sql
CREATE TABLE post_tags (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  post_id BIGINT NOT NULL,
  tag_id BIGINT NOT NULL,
  FOREIGN KEY (post_id) REFERENCES posts(id),
  FOREIGN KEY (tag_id) REFERENCES tags(id),
  UNIQUE KEY (post_id, tag_id)
);
```

---

## 🔍 Where is the FK Column?

| Relation Type | FK Column Location | Field Name |
|---------------|-------------------|------------|
| **belongsTo** | Source entity table | `source_field_name` at source |
| **hasOne** | Target entity table ⚠️ | `source_field_name` at target |
| **hasMany** | Target entity table ⚠️ | `source_field_name` at target |
| **manyToMany** | Junction table | Both FKs in junction |

**Key Insight:**
- `source_field_name` = Field name that will be used
- Location depends on `relation_type`
- For hasOne/hasMany: field physically exists in **target entity**

---

## 🎨 API/DTO Mapping

### **belongsTo Example:**

**Relation:**
```
Post belongsTo User (field: user_id)
```

**DTO (CreatePostRequest):**
```go
type CreatePostRequest struct {
    Title    string `json:"title"`
    UserUUID string `json:"user_uuid"`  // API uses UUID
}
```

**Model (Post):**
```go
type Post struct {
    ID     int64  `db:"id"`
    Title  string `db:"title"`
    UserID int64  `db:"user_id"`  // Internal ID
}
```

**Handler Translation:**
```go
// UUID → ID
userUUID, _ := uuid.Parse(req.UserUUID)
var userID int64
db.Get(&userID, "SELECT id FROM users WHERE uuid = ?", userUUID)
post.UserID = userID
```

---

### **hasMany Example:**

**Relation:**
```
Product hasMany Stock (field: product_id)
```

**DTO (CreateStockRequest):**
```go
type CreateStockRequest struct {
    ProductUUID  string `json:"product_uuid"`  // API uses UUID
    LocationUUID string `json:"location_uuid"`
    Quantity     int    `json:"quantity"`
}
```

**Model (Stock):**
```go
type Stock struct {
    ID         int64 `db:"id"`
    ProductID  int64 `db:"product_id"`   // FK exists in Stock
    LocationID int64 `db:"location_id"`  // FK exists in Stock
    Quantity   int   `db:"quantity"`
}
```

---

## ⚙️ Code Generation Logic

### **1. Determine FK Location:**

```go
func getFKLocation(relationType string, entityID int64, relation Relation) (hasFK bool, targetTable string) {
    switch relationType {
    case "belongsTo":
        // FK in source entity
        if relation.SourceEntityID == entityID {
            return true, relation.TargetEntityName
        }
    case "hasOne", "hasMany":
        // FK in target entity
        if relation.TargetEntityID == entityID {
            return true, relation.SourceEntityName
        }
    case "manyToMany":
        // FK in junction table
        return false, ""
    }
    return false, ""
}
```

### **2. Generate Migration:**

```go
func generateMigration(entity Entity, relations []Relation) string {
    // Regular fields
    for field in entity.Fields {
        if !isFKField(field.Name, relations) {
            sql += field.Name + " " + field.Type
        }
    }
    
    // FK fields from relations
    for rel in relations {
        if shouldAddFK(entity.ID, rel) {
            fkColumn := rel.SourceFieldName
            targetTable := getTargetTable(rel)
            sql += fkColumn + " BIGINT NOT NULL"
            sql += "FOREIGN KEY (" + fkColumn + ") REFERENCES " + targetTable + "(id)"
        }
    }
}
```

---

## 🧪 Testing Scenarios

### **Test Case 1: Product hasMany Stock**

**Setup:**
```sql
INSERT INTO entities (id, name, table_name) VALUES
  (1, 'Product', 'products'),
  (2, 'Stock', 'stocks');

INSERT INTO relations (
  source_entity_id, 
  target_entity_id, 
  source_field_name, 
  relation_type
) VALUES (
  1,              -- Product (source)
  2,              -- Stock (target)
  'product_id',   -- FK field name
  'hasMany'
);
```

**Expected Migration (stocks table):**
```sql
CREATE TABLE stocks (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  uuid CHAR(36) UNIQUE NOT NULL,
  quantity INT NOT NULL,
  product_id BIGINT NOT NULL,  -- ✅ Added from relation
  FOREIGN KEY (product_id) REFERENCES products(id)
);
```

**Expected API:**
```bash
POST /stocks
{
  "product_uuid": "550e8400-...",  # User sends UUID
  "quantity": 100
}

# Handler translates UUID → ID
# Inserts with product_id = internal_id
```

---

### **Test Case 2: Post belongsTo User**

**Setup:**
```sql
INSERT INTO relations (
  source_entity_id, 
  target_entity_id, 
  source_field_name, 
  relation_type
) VALUES (
  3,           -- Post (source)
  4,           -- User (target)
  'user_id',   -- FK field name
  'belongsTo'
);
```

**Expected Migration (posts table):**
```sql
CREATE TABLE posts (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  uuid CHAR(36) UNIQUE NOT NULL,
  title VARCHAR(255) NOT NULL,
  user_id BIGINT NOT NULL,  -- ✅ Added from relation
  FOREIGN KEY (user_id) REFERENCES users(id)
);
```

---

## 📊 Consistency Matrix

| Element | belongsTo | hasOne | hasMany | manyToMany |
|---------|-----------|---------|---------|------------|
| **source_entity_id** | Entity yang belongs | Entity yang has | Entity yang has | Entity A |
| **target_entity_id** | Entity yang dirujuk | Entity yang dimiliki | Entity yang dimiliki | Entity B |
| **source_field_name** | FK at source | FK at target | FK at target | N/A (junction) |
| **FK Location** | Source table | Target table | Target table | Junction table |
| **API Field** | `{target}_uuid` | N/A | N/A | N/A |

---

## 🎯 Best Practices

### **1. Naming Convention:**
- Always use `{entity}_id` format for FK fields
- Example: `user_id`, `product_id`, `category_id`
- Consistent with snake_case
- Custom names allowed: `author_id`, `owner_id`, `parent_id`

### **2. Field Name Selection:**
User can customize FK field name, but must follow rules:
- ✅ **Allowed**: `user_id`, `author_id`, `owner_id`, `created_by_id`
- ⚠️ **Caution**: `user`, `author` (missing _id suffix, but still allowed)
- ❌ **Not Recommended**: `userId`, `UserID` (not snake_case)
- ❌ **Conflict**: Field name that exists with different type

### **3. Type Consistency:**
- FK field is **always BIGINT** (matches PK type)
- User only customizes field **name**, not type
- If entity has existing field with same name:
  - If type is `int` or `int64`: OK (will be overridden to BIGINT)
  - If type is other (string, bool, etc): **ERROR - conflict**

### **4. Validation Rules:**
- ✅ Minimum 3 characters
- ✅ No conflict with existing non-FK fields
- ✅ Snake_case format (auto-converted)
- ✅ Validates in correct entity based on relation type

### **5. API Consistency:**
- Always expose as `{entity}_uuid` in API
- Example: `user_uuid`, `product_uuid`
- Never expose internal IDs

### **6. Validation:**
- Check if target entity exists before creating relation
- Validate FK field name doesn't conflict with existing fields
- Ensure target table exists in database

### **7. Documentation:**
- Comment in code which entity has FK
- Example: `// FK product_id exists in stocks table (target)`

---

## 🔧 Migration Template

```go
// For hasMany/hasOne relations
if relation.Type == "hasMany" || relation.Type == "hasOne" {
    // FK is in TARGET entity
    if entity.ID == relation.TargetEntityID {
        fkColumn := relation.SourceFieldName
        sourceTable := getEntityTableName(relation.SourceEntityID)
        
        sql += fkColumn + " BIGINT NOT NULL"
        sql += "FOREIGN KEY (" + fkColumn + ") REFERENCES " + sourceTable + "(id)"
        
        if relation.Type == "hasOne" {
            sql += "UNIQUE KEY (" + fkColumn + ")"
        }
    }
}

// For belongsTo relations
if relation.Type == "belongsTo" {
    // FK is in SOURCE entity
    if entity.ID == relation.SourceEntityID {
        fkColumn := relation.SourceFieldName
        targetTable := getEntityTableName(relation.TargetEntityID)
        
        sql += fkColumn + " BIGINT NOT NULL"
        sql += "FOREIGN KEY (" + fkColumn + ") REFERENCES " + targetTable + "(id)"
    }
}
```

---

## ❓ FAQ

**Q: Why is it called `source_field_name` even when FK is in target entity?**
A: For API consistency. We always describe relation from source perspective:
- "Product hasMany Stock via product_id"
- Even though `product_id` physically exists in `stocks` table

**Q: Can I customize FK field name?**
A: Yes! Set `source_field_name` when creating relation. Default is `{source_entity}_id`.
Examples: `author_id`, `owner_id`, `created_by_id`, `parent_id` - any name is allowed!

**Q: What if FK field name conflicts with existing field?**
A: Validation checks for conflicts:
- If existing field is `int` or `int64` → OK (will be overridden to BIGINT)
- If existing field is other type (string, bool, etc) → ERROR (must choose different name)
- Error message shows which entity has the conflict

**Q: Can I use field names without '_id' suffix?**
A: Yes, but not recommended. Examples like `owner` or `author` work, but `owner_id` is clearer.
The system validates name length (min 3 chars) but allows any format.

**Q: What type will the FK field be?**
A: Always `BIGINT NOT NULL` (matches primary key type). 
User only customizes the field NAME, not the type.

**Q: What if I want bidirectional relation?**
A: Create TWO relations:
- A hasMany B (creates FK in B)
- B belongsTo A (documents reverse relation)

**Q: How to handle composite keys?**
A: Not supported yet. Use manyToMany with junction table.

**Q: Can I change relation type after creation?**
A: Requires re-deployment. Old FK will be dropped, new one created.

**Q: What happens if I delete entity with relations?**
A: Relations are automatically deleted (cascade delete in relations table).
Generated service uses FK constraints with ON DELETE behavior you specify.

---

## 📝 Summary

**Key Takeaways:**

1. ✅ **source_entity_id** = Entity that "owns" the relation
2. ✅ **target_entity_id** = Entity that is referenced
3. ✅ **source_field_name** = FK field name (location depends on type)
4. ✅ For **belongsTo**: FK in source entity
5. ✅ For **hasOne/hasMany**: FK in **target entity** ⚠️
6. ✅ For **manyToMany**: FK in junction table
7. ✅ API always uses **UUID** (`{entity}_uuid`)
8. ✅ Internal model uses **ID** (`{entity}_id`)

**The naming might seem confusing, but it's consistent with ActiveRecord/Laravel conventions!**

---

*Last Updated: January 28, 2026*  
*Version: 1.0*
