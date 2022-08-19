idgen
====

    全局型 ID 生成


```go

g := NewWithShard(0)

// Generate new global id
newId := g.Next()

// Or generate with special time
tm := time.Now()
nextId := g.NextWithTime(tm)

```
