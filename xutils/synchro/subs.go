package synchro

// SyncSubs 同步变更的数据
type Subs struct {
	Exp     []SubExp    `json:",omitempty"`
	Roles   []Role      `json:",omitempty"`
	Equips  []Equip     `json:",omitempty"`
	Objects []SubObject `json:",omitempty"`
	Tasks   []Task      `json:",omitempty"`
	dsm     bool
}

// DisableMerge 关闭合并功能
func (o *Subs) DisableMerge() {
	o.dsm = true
}

// EnableMerge 开启合并功能
func (o *Subs) EnableMerge() {
	o.dsm = false
}

// SubExp 经验同步
type SubExp struct {
	Level int32
	Cur   int32
	Max   int32
	Sub   int32
}

// SyncExp 同步经验值
func (o *Subs) SyncExp(level, cur, max, sub int32) {
	if len(o.Exp) == 0 {
		o.Exp = append(o.Exp, SubExp{
			Level: level, Cur: cur, Max: max, Sub: sub,
		})
	} else {
		o.Exp[0].Level = level
		o.Exp[0].Cur = cur
		o.Exp[0].Max = max
		o.Exp[0].Sub += sub
	}
}

// Role 玩家角色
type Role struct {
	Id     string  // 角色序号
	Rid    string  // 角色ID
	Level  int32   // 等级
	Star   int32   // 星级
	Awaken int32   // 觉醒次数
	Skl    []int32 // 技能等级
}

// SyncRole 同步角色
func (o *Subs) SyncRole(r *Role) {
	for i := 0; i < len(o.Roles); i++ {
		if o.Roles[i].Id == r.Id {
			o.Roles[i].Rid = r.Rid
			o.Roles[i].Level = r.Level
			o.Roles[i].Star = r.Star
			o.Roles[i].Awaken = r.Awaken
			o.Roles[i].Skl = r.Skl
			return
		}
	}
	o.Roles = append(o.Roles, *r)
}

// Equip 装备
type Equip struct {
	Id      string  // 装备序号
	Eid     string  // 装备ID
	Level   int32   // 等级
	Star    int32   // 星级
	Refined int32   // 精练等级
	Pis     []int32 // 拥有的属性值
}

// SyncEquip 同步装备
func (o *Subs) SyncEquip(eq *Equip) {
	for i := 0; i < len(o.Equips); i++ {
		if o.Equips[i].Id == eq.Id {
			o.Equips[i].Eid = eq.Eid
			o.Equips[i].Level = eq.Level
			o.Equips[i].Star = eq.Star
			o.Equips[i].Refined = eq.Refined
			o.Equips[i].Pis = eq.Pis
			return
		}
	}
	o.Equips = append(o.Equips, *eq)
}

// SubObject 物品变化
type SubObject struct {
	Id  string
	Num int32
	Sub int32
}

// SyncObject 同步物品
func (o *Subs) SyncObject(id string, num, sub int32) {
	if !o.dsm {
		for i := 0; i < len(o.Objects); i++ {
			if o.Objects[i].Id == id {
				o.Objects[i].Num = num
				o.Objects[i].Sub += sub
				return
			}
		}
	}
	o.Objects = append(o.Objects, SubObject{
		Id: id, Num: num, Sub: sub,
	})
}

// Task 任务
type Task struct {
	Id    string
	State int32
}

// SyncTask 同步任务
func (o *Subs) SyncTask(id string, state int32) {
	for i := 0; i < len(o.Tasks); i++ {
		if o.Tasks[i].Id == id {
			o.Tasks[i].State = state
			return
		}
	}
	o.Tasks = append(o.Tasks, Task{
		Id: id, State: state,
	})
}
