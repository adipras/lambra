# Testing Checklist - Visual Relations Feature

## Pre-Testing Setup

- [ ] Docker services running (`make up`)
- [ ] Database migrations applied (`make migrate-up`)
- [ ] Backend compiled successfully
- [ ] Frontend built successfully
- [ ] At least 2 test entities created

---

## Phase 1: Basic Relation Creation (Drag-and-Drop)

### Test Case 1.1: Create belongsTo Relation
- [ ] Open Diagram View
- [ ] Drag from Entity A to Entity B
- [ ] RelationModal opens
- [ ] Select "Belongs To" relation type
- [ ] Verify field name auto-generated (e.g., `entity_b_id`)
- [ ] Set ON DELETE to CASCADE
- [ ] Click "Create Relation"
- [ ] Verify pink edge appears between entities
- [ ] Verify edge label shows "→ belongs to"

**Expected Result:** ✅ belongsTo relation created successfully

### Test Case 1.2: Create hasMany Relation
- [ ] Drag from Entity A to Entity B
- [ ] Select "Has Many" relation type
- [ ] Verify field name auto-generated
- [ ] Set ON DELETE to SET NULL
- [ ] Click "Create Relation"
- [ ] Verify blue edge appears
- [ ] Verify edge label shows "1:N has many"

**Expected Result:** ✅ hasMany relation created successfully

### Test Case 1.3: Create manyToMany Relation
- [ ] Drag from Entity A to Entity B
- [ ] Select "Many to Many" relation type
- [ ] Verify junction table name auto-generated (e.g., `entity_as_entity_bs`)
- [ ] Click "Create Relation"
- [ ] Verify orange edge appears
- [ ] Verify edge label shows "N:M many to many"

**Expected Result:** ✅ manyToMany relation created successfully

---

## Phase 2: Edge Interaction & Deletion

### Test Case 2.1: Hover Over Edge
- [ ] Hover over a relation edge
- [ ] Verify tooltip appears with:
  - [ ] Relation type (uppercase)
  - [ ] Field name
  - [ ] ON DELETE behavior (color-coded)
- [ ] Verify edge scales up (1.0 → 1.1)
- [ ] Verify edge thickness increases (2px → 3px)

**Expected Result:** ✅ Tooltip and visual feedback working

### Test Case 2.2: Delete Relation
- [ ] Hover over a relation edge
- [ ] Click trash icon that appears
- [ ] Verify confirmation dialog appears
- [ ] Click "OK"
- [ ] Verify edge disappears
- [ ] Verify relation deleted from database

**Expected Result:** ✅ Relation deleted successfully

---

## Phase 3: Code Generation with Relations

### Test Case 3.1: Generate Code with belongsTo
- [ ] Create 2 entities (User, Post)
- [ ] Create belongsTo relation: Post → User
- [ ] Go to "Generate" or deploy
- [ ] Check generated migration SQL
- [ ] Verify FK column created (`user_id BIGINT`)
- [ ] Verify FOREIGN KEY constraint added
- [ ] Verify ON DELETE CASCADE applied

**Expected Result:** ✅ Generated code includes FK constraints

### Test Case 3.2: Generate Code with manyToMany
- [ ] Create 2 entities (Post, Tag)
- [ ] Create manyToMany relation: Post ↔ Tag
- [ ] Generate code
- [ ] Verify junction table created (`posts_tags`)
- [ ] Verify junction table has both FKs (`post_id`, `tag_id`)
- [ ] Verify PRIMARY KEY (post_id, tag_id)

**Expected Result:** ✅ Junction table generated correctly

---

## Phase 4: Deployment & Database Verification

### Test Case 4.1: Deploy Service with Relations
- [ ] Create project with entities and relations
- [ ] Click "Deploy"
- [ ] Wait for deployment to complete
- [ ] Verify service starts successfully
- [ ] Check deployment logs for migration success

**Expected Result:** ✅ Service deployed with FK constraints

### Test Case 4.2: Verify FK Constraints in Database
- [ ] Connect to deployed service database
- [ ] Run: `SHOW CREATE TABLE <table_name>;`
- [ ] Verify FOREIGN KEY constraints exist
- [ ] Verify ON DELETE behavior matches configuration

```sql
-- Example verification query
SHOW CREATE TABLE posts;
-- Should show: FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
```

**Expected Result:** ✅ FK constraints applied correctly

---

## Phase 5: Snapshot & Rollback with Relations

### Test Case 5.1: Create Snapshot with Relations
- [ ] Create entities with relations
- [ ] Deploy project
- [ ] Verify snapshot created automatically
- [ ] Check snapshot metadata includes relations

**Expected Result:** ✅ Snapshot includes relations data

### Test Case 5.2: Rollback to Previous Snapshot
- [ ] Modify relations (add/delete)
- [ ] Deploy again
- [ ] Go to Snapshots
- [ ] Click "Rollback" on previous snapshot
- [ ] Verify relations restored to previous state
- [ ] Check diagram shows correct relations

**Expected Result:** ✅ Relations restored on rollback

---

## Phase 6: Edge Cases & Error Handling

### Test Case 6.1: Prevent Self-Reference (if implemented)
- [ ] Try to drag from Entity A to Entity A
- [ ] Verify error or prevention mechanism
- [ ] (Or verify self-reference works if intended)

**Expected Result:** ✅ Handled correctly per spec

### Test Case 6.2: Delete Entity with Relations
- [ ] Create Entity A with relations to Entity B
- [ ] Delete Entity A
- [ ] Verify relations cascade deleted (soft delete)
- [ ] Verify edges disappear from diagram

**Expected Result:** ✅ Relations cascade deleted

### Test Case 6.3: Multiple Relations Between Same Entities
- [ ] Create Entity User and Entity Organization
- [ ] Create relation: User belongsTo Organization (organization_id)
- [ ] Create relation: User belongsTo Organization as Manager (manager_id)
- [ ] Verify both relations exist
- [ ] Verify different field names

**Expected Result:** ✅ Multiple relations allowed

---

## Phase 7: Backward Compatibility

### Test Case 7.1: Legacy Field-based Relations
- [ ] If migration script NOT run, old relations should show
- [ ] Open Diagram View
- [ ] Verify legacy relations show as dashed lines
- [ ] Verify 50% opacity
- [ ] Verify "Legacy" indicator in tooltip
- [ ] Verify no delete button for legacy relations

**Expected Result:** ✅ Legacy relations displayed correctly

### Test Case 7.2: Mixed Old and New Relations
- [ ] Project with both legacy and new relations
- [ ] Verify both types display
- [ ] Verify visual distinction (dashed vs solid)
- [ ] Verify only new relations can be deleted

**Expected Result:** ✅ Both formats coexist peacefully

---

## Phase 8: Performance Testing

### Test Case 8.1: Large Diagram (20+ Entities)
- [ ] Create 20+ entities
- [ ] Create 50+ relations
- [ ] Open Diagram View
- [ ] Verify rendering performance acceptable (< 2s initial load)
- [ ] Test zoom in/out responsiveness
- [ ] Test drag-and-drop performance

**Expected Result:** ✅ Performance acceptable

### Test Case 8.2: API Response Time
- [ ] Use browser DevTools Network tab
- [ ] Monitor `/entities/:id/relations` API calls
- [ ] Verify response time < 500ms
- [ ] Monitor memory usage (< 100MB for diagram)

**Expected Result:** ✅ API performance acceptable

---

## Phase 9: UI/UX Verification

### Test Case 9.1: RelationModal UX
- [ ] Open RelationModal
- [ ] Verify all fields visible and accessible
- [ ] Verify color-coded relation type cards
- [ ] Verify helper text displays
- [ ] Verify validation errors show correctly
- [ ] Test keyboard navigation (Tab)
- [ ] Test Escape key to close modal

**Expected Result:** ✅ Modal UX smooth and intuitive

### Test Case 9.2: Info Panel & Instructions
- [ ] Check bottom-left info panel
- [ ] Verify instructions clear
- [ ] Verify relation color legend visible
- [ ] Verify drag-to-connect instruction present

**Expected Result:** ✅ Instructions clear and helpful

### Test Case 9.3: EntityForm Info Banner
- [ ] Open EntityForm
- [ ] Verify info banner shows between Entity Info and Fields
- [ ] Verify gradient background (pink-purple)
- [ ] Verify color indicators present
- [ ] Verify "relation" NOT in field type dropdown

**Expected Result:** ✅ Banner guides users to Diagram View

---

## Summary Checklist

### Critical Path (Must Pass)
- [ ] Create relation via drag-and-drop
- [ ] Relations display as colored edges
- [ ] Delete relation works
- [ ] Generated code includes FK constraints
- [ ] Deployment works with relations
- [ ] Snapshot/rollback includes relations

### Important (Should Pass)
- [ ] Edge hover tooltips work
- [ ] Multiple relation types work
- [ ] Legacy relations display correctly
- [ ] EntityForm has no relation option
- [ ] Info banner guides users

### Nice to Have (Optional)
- [ ] Performance with 50+ relations
- [ ] Self-referencing relations
- [ ] Multiple relations between entities
- [ ] Migration script works

---

## Bug Report Template

If you find issues, report using this format:

```
**Bug Title:** [Brief description]

**Severity:** Critical / High / Medium / Low

**Steps to Reproduce:**
1. 
2. 
3. 

**Expected Result:**
[What should happen]

**Actual Result:**
[What actually happened]

**Screenshots/Logs:**
[If applicable]

**Environment:**
- Backend: [running/not running]
- Frontend: [running/not running]
- Database: [version]
- Browser: [name/version]
```

---

## Sign-off

**Tested By:** _______________  
**Date:** _______________  
**Result:** ✅ PASS / ❌ FAIL  
**Notes:** _______________

---

**Last Updated:** 2026-01-22  
**Version:** 1.0
