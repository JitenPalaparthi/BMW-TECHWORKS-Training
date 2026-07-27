
let places = new Map();
//let fruits= []//new Array();

places.set(560086,"Bangalore-1")
places.set(560096,"Bangalore-2")
places.set(560034,"Bangalore-3")
places.set(522001,"Guntur-1")
places.set(522002,"Guntur-2")
places.set(695011,"Trv-1")
places.set(1010,"no place");

console.log(places)

if (places.has(560036)==true){
   let place= places.get(560036);
   console.log(place); 
}else{
    console.log("Key does not exist");
}


// console.log(places.get(1));

let place= places.get(560086);// if key does not exist , it is undefined

console.log(place);


places.set(695011,"Thiruvanantapuram-1"); // update a value, unsert(if the key does not exist, it does the insert operation)


let done=places.delete(1010);

if (done==true){
    console.log("The key existed, and deleted");
}else{
    console.log("key does not exist, so nothing to delete");
}

// to fetch all the keys and values from the map 

for(let element of places){
    console.log("===>",element);
}

console.log();
for (let [key,value] of places){
    console.log("Key:",key,"Value:",value);
}

// get only keys 

for (let key of places.keys()){
    console.log(key)
}

// write a functuon to fetch only values 

// To use for each loop 

places.forEach(function(key,value){
console.log("Key:",key,"Value:",value);
});


// java Script, maps are more advanced so that can store not only just a value but also functions

