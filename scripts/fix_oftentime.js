// MongoDB migration: fix OftenTime fields stored as empty document {}
// Run: mongosh studygolang scripts/fix_oftentime.js

var now = new Date();

var collections = [
  { name: "topics",         fields: ["ctime", "mtime", "lastreplytime"] },
  { name: "articles",       fields: ["ctime", "mtime", "lastreplytime"] },
  { name: "comments",       fields: ["ctime"] },
  { name: "resource",       fields: ["ctime", "mtime", "lastreplytime"] },
  { name: "open_project",   fields: ["ctime", "mtime", "lastreplytime"] },
  { name: "books",          fields: ["created_at", "updated_at", "lastreplytime"] },
  { name: "wiki",           fields: ["ctime"] },
  { name: "user_info",      fields: ["ctime"] },
  { name: "morning_reading", fields: ["ctime"] },
  { name: "message",        fields: ["ctime"] },
  { name: "system_message", fields: ["ctime"] },
  { name: "subject",        fields: ["created_at", "updated_at"] },
  { name: "feed",           fields: ["created_at", "updated_at", "lastreplytime"] },
];

collections.forEach(function(col) {
  var coll = db.getCollection(col.name);
  col.fields.forEach(function(field) {
    var filter = {};
    filter[field] = { $type: "object" };
    var update = { $set: {} };
    update.$set[field] = now;
    var result = coll.updateMany(filter, update);
    if (result.modifiedCount > 0) {
      print("[" + col.name + "." + field + "] fixed " + result.modifiedCount + " docs (was empty object)");
    }
  });
});

print("\nDone!");
