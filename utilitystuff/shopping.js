function shopCMD(prompt, chatObj) {
    prompt = prompt.trim().split(/\s+/).slice(1).join(' ').trim();

    const quantity = prompt.match(/\d+/);
    if (quantity) {
        prompt = prompt.replace(new RegExp("\\s*" + quantity[0] + "\\s*"), '').trim();
    }
    
    const itemToBuy = prompt

    if (!itemToBuy) {
        return "Usage: !shop <item> or !shop clear";
    }

    if (itemToBuy==="clear") {
        chatObj.shoppingList = [];
        return "Shopping list cleared.";
    }

    if (!chatObj.shoppingList) {
        chatObj.shoppingList = [];
    }

    const existingItem = chatObj.shoppingList.find(item =>
        item.item.toLowerCase() === itemToBuy.toLowerCase()
    );

    if (existingItem) {
        if (quantity) {
            existingItem.quantity = parseInt(quantity[0]);
            return `"${itemToBuy}" quantity updated to ${existingItem.quantity}.`;
        } else {
            return `"${itemToBuy}" is already in the shopping list.`;
        }
    }

    if (quantity) {
        chatObj.shoppingList.push({ item: itemToBuy, quantity: parseInt(quantity[0]) });
    } else {
        chatObj.shoppingList.push({ item: itemToBuy, quantity: 0 });
    }

    return ((quantity?quantity[0]+" x ":"")+'"'+itemToBuy+`" ${quantity&&quantity>0?"have":"has"} been added to shopping list`);
}

function shopListCMD(chatObj) {
    if (!chatObj.shoppingList || chatObj.shoppingList.length === 0) {
        return "Your shopping list is empty.";
    }

    let list = "Shopping list:\n";
    chatObj.shoppingList.forEach((item, index) => {
        list += `  ${index + 1}.  ${item.quantity>0?item.quantity+" x ":""}${item.item}\n`;
    });

    return list.trim();
}

function unshopCMD(prompt, chatObj) {
    const itemToRemove = prompt.trim().split(/\s+/).slice(1).join(' ').trim();

    if (!itemToRemove) {
        return "Usage: !unshop <item>";
    }

    if (!chatObj.shoppingList || chatObj.shoppingList.length === 0) {
        return "Your shopping list is empty.";
    }

    const itemIndex = itemToRemove.match(/\d+/);

    if (itemIndex) {
        const index = parseInt(itemIndex[0]) - 1;
        if (index < chatObj.shoppingList.length) {
            const removedItem = chatObj.shoppingList.splice(index, 1)[0];
            return `"${removedItem.item}" has been removed from your shopping list.`;
        } else {
            return "Invalid item index.";
        }
    }

    for (let i = 0; i < chatObj.shoppingList.length; i++) {
        if (chatObj.shoppingList[i].item.match(new RegExp(itemToRemove, "i"))) {
            const removedItem = chatObj.shoppingList.splice(i, 1)[0];
            return `"${removedItem.item}" has been removed from your shopping list.`;
        }
    }

    return `Item "${itemToRemove}" not found in shopping list.`;
}

module.exports = { shopCMD, unshopCMD, shopListCMD}