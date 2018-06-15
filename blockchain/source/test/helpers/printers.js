export const inspect = async (data, clazz) => {
    for (let k in clazz) {
        console.log(`${k}: ${data[clazz[k]]} `);
    }
};
