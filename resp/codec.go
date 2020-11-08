package resp

import "github.com/micro/packet"

// Decode RespSubs generate by codec.
func (o *SyncSubs) Decode(p *packet.Packet) {
	cExp := p.ReadU64()
	o.Exp = make([]SubExp, cExp)
	for i := uint64(0); i < cExp; i++ {
		o.Exp[i].Decode(p)
	}
	cRoles := p.ReadU64()
	o.Roles = make([]Role, cRoles)
	for i := uint64(0); i < cRoles; i++ {
		o.Roles[i].Decode(p)
	}
	cEquips := p.ReadU64()
	o.Equips = make([]Equip, cEquips)
	for i := uint64(0); i < cEquips; i++ {
		o.Equips[i].Decode(p)
	}
	cObjects := p.ReadU64()
	o.Objects = make([]SubObject, cObjects)
	for i := uint64(0); i < cObjects; i++ {
		o.Objects[i].Decode(p)
	}
	cTasks := p.ReadU64()
	o.Tasks = make([]Task, cTasks)
	for i := uint64(0); i < cTasks; i++ {
		o.Tasks[i].Decode(p)
	}
}

// Encode RespSubs generate by codec.
func (o *SyncSubs) Encode(p *packet.Packet) {
	cExp := uint64(len(o.Exp))
	p.WriteU64(cExp)
	for i := uint64(0); i < cExp; i++ {
		o.Exp[i].Encode(p)
	}
	cRoles := uint64(len(o.Roles))
	p.WriteU64(cRoles)
	for i := uint64(0); i < cRoles; i++ {
		o.Roles[i].Encode(p)
	}
	cEquips := uint64(len(o.Equips))
	p.WriteU64(cEquips)
	for i := uint64(0); i < cEquips; i++ {
		o.Equips[i].Encode(p)
	}
	cObjects := uint64(len(o.Objects))
	p.WriteU64(cObjects)
	for i := uint64(0); i < cObjects; i++ {
		o.Objects[i].Encode(p)
	}
	cTasks := uint64(len(o.Tasks))
	p.WriteU64(cTasks)
	for i := uint64(0); i < cTasks; i++ {
		o.Tasks[i].Encode(p)
	}
}


// Decode SubExp generate by codec.
func (o *SubExp) Decode(p *packet.Packet) {
	o.Level = p.ReadI32()
	o.Cur = p.ReadI32()
	o.Max = p.ReadI32()
	o.Sub = p.ReadI32()
}

// Encode SubExp generate by codec.
func (o *SubExp) Encode(p *packet.Packet) {
	p.WriteI32(o.Level)
	p.WriteI32(o.Cur)
	p.WriteI32(o.Max)
	p.WriteI32(o.Sub)
}


// Decode Role generate by codec.
func (o *Role) Decode(p *packet.Packet) {
	o.Id = p.ReadString()
	o.Rid = p.ReadString()
	o.Level = p.ReadI32()
	o.Star = p.ReadI32()
	o.Awaken = p.ReadI32()
	o.Skl = p.ReadI32S()
}

// Encode Role generate by codec.
func (o *Role) Encode(p *packet.Packet) {
	p.WriteString(o.Id)
	p.WriteString(o.Rid)
	p.WriteI32(o.Level)
	p.WriteI32(o.Star)
	p.WriteI32(o.Awaken)
	p.WriteI32S(o.Skl)
}


// Decode Equip generate by codec.
func (o *Equip) Decode(p *packet.Packet) {
	o.Id = p.ReadString()
	o.Eid = p.ReadString()
	o.Level = p.ReadI32()
	o.Star = p.ReadI32()
	o.Refined = p.ReadI32()
	o.Pis = p.ReadI32S()
}

// Encode Equip generate by codec.
func (o *Equip) Encode(p *packet.Packet) {
	p.WriteString(o.Id)
	p.WriteString(o.Eid)
	p.WriteI32(o.Level)
	p.WriteI32(o.Star)
	p.WriteI32(o.Refined)
	p.WriteI32S(o.Pis)
}


// Decode SubObject generate by codec.
func (o *SubObject) Decode(p *packet.Packet) {
	o.Id = p.ReadString()
	o.Num = p.ReadI32()
	o.Sub = p.ReadI32()
}

// Encode SubObject generate by codec.
func (o *SubObject) Encode(p *packet.Packet) {
	p.WriteString(o.Id)
	p.WriteI32(o.Num)
	p.WriteI32(o.Sub)
}


// Decode Task generate by codec.
func (o *Task) Decode(p *packet.Packet) {
	o.Id = p.ReadString()
	o.State = p.ReadI32()
}

// Encode Task generate by codec.
func (o *Task) Encode(p *packet.Packet) {
	p.WriteString(o.Id)
	p.WriteI32(o.State)
}

