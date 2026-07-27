
let people  = new Map();

people.set(101,{"name":"Jiten","Sal":100000});
people.set(102,{"name":"Raja","Sal":34324});
people.set(103,{"name":"Priya","Sal":42343242});
people.set(104,{"name":"David","Sal":3423423});
people.set(105,{"name":"John","Sal":2343423});

let person = {
  "name":"Jiten",
  "sal":324324,
  "greet":function(){
    console.log("Hello World!");
  },
  "address":{
    "line-1":"near medical college",
    "city":"Trivandrum",
    "state":"Kerala",
    "just":function(){
        console.log("do something");
    },
    "mymap":people,
    "socialMediaq":["linkedin","x","fb"]
  }
};


console.log(person);

person.greet();
console.log(person.name);
console.log(person.sal);
console.log(person.age);





for(let person of people){
    console.log("ID:",person[0],"Name:",person[1].name,"Sal:",person[1].Sal);
}



let operations = new Map();

operations.set("add",(a,b)=>a+b);
operations.set("sub",(a,b)=>a-b);
operations.set("mul",(a,b)=>a*b);
operations.set("div",(a,b)=>a/b);
console.log(typeof operations);