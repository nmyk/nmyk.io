const words = {
    "adverbs": ["afterward", "again", "ahead", "alike", "alone", "already", "also", "anxiously", "anyway", "anywhere", "apart", "around", "aside", "asynchronously", "automatically", "away", "back", "badly", "better", "carefully", "certainly", "clearly", "closely", "completely", "concurrently", "constantly", "correctly", "daily", "directly", "double", "doubtfully", "down", "dutifully", "earlier", "early", "earnestly", "easily", "enough", "entirely", "equally", "eternally", "eventually", "everywhere", "exactly", "fairly", "faithfully", "far", "fast", "finally", "first", "for-real", "forever", "forward", "frequently", "fully", "further", "furtively", "generally", "gently", "gradually", "greatly", "guiltily", "happily", "heartily", "here", "hopelessly", "however", "immediately", "indeed", "instead", "intensely", "internally", "ironically", "irresistably", "later", "long", "loudly", "lovingly", "luckily", "madly", "masterfully", "maybe", "more", "mostly", "naturally", "nearby", "neatly", "never", "nevertheless", "nicely", "now", "obsessively", "occasionally", "once", "out", "particularly", "partly", "perfectly", "perhaps", "prettily", "properly", "quickly", "quietly", "randomly", "rapidly", "recently", "repeatedly", "sarcastically", "silly", "simply", "slightly", "sloppily", "slowly", "softly", "somehow", "sometime", "somewhere", "soon", "still", "suddenly", "there", "therefore", "thus", "tightly", "together", "twice", "unstoppably", "up", "upward", "usually", "viciously", "well", "widely", "with-me", "within"],
    "gerunds": ["acting", "backing-up", "baking", "bathing", "beaming", "becoming", "beginning", "blinking", "boiling", "bowling", "bragging", "bullying", "bursting", "buzzing", "changing", "chatting", "cheering", "cleaning", "cleaving", "coding", "collaborating", "connecting", "contracting", "cooking", "covering", "crushing", "cutting", "dating", "daydreaming", "disgracing", "disrupting", "dissolving", "downloading", "drinking", "driving", "eating", "enfolding", "exercising", "expanding", "exploding", "fighting", "flailing", "fleeing", "flipping", "floating", "flying", "freewheeling", "gardening", "getting", "giving", "glistening", "glowing", "grinning", "healing", "hiding", "honoring", "hoping", "humiliating", "hunting", "ignoring", "inflaming", "ingesting", "insulting", "integrating", "investigating", "joking", "laughing", "leaping", "leaving", "leering", "looming", "losing", "loving", "lying", "making", "missing", "moping", "moving", "napping", "oiling", "panicking", "playing", "pointing", "pouring", "preaching", "protesting", "putting", "quilting", "recycling", "relaxing", "respecting", "revolutionizing", "revolving", "riding", "rising", "rotating", "running", "saving", "scrubbing", "seeing", "setting", "shaking", "shining", "shivering", "shopping", "singing", "sitting", "skiing", "sleeping", "slicing", "smoking", "sneering", "snoring", "sobbing", "solving", "stashing", "stinging", "studying", "sweating", "swimming", "swinging", "thinking", "tinkering", "toiling", "training", "transforming", "trying", "typing", "understanding", "using", "vibrating", "weeping", "winning", "wishing", "wobbling", "worrying", "worshipping", "wringing", "writing", "yearning"],
    "adjectives": ["able", "active", "actual", "additional", "ambient", "angry", "animated", "atmospheric", "atomic", "available", "average", "bad", "basic", "best", "bigger", "biggest", "busy", "central", "certain", "characteristic", "clear", "comfortable", "common", "complete", "complex", "concerned", "continued", "curious", "current", "daily", "dangerous", "dead", "depressed", "different", "difficult", "direct", "due", "easier", "easy", "electric", "empty", "entire", "equal", "essential", "extra", "familiar", "famous", "feline", "final", "finest", "flat", "foreign", "former", "fourth", "free", "fresh", "friendly", "full", "furry", "general", "golden", "good", "great", "greater", "greatest", "green", "happy", "hard", "heavy", "high", "higher", "highest", "hot", "huge", "important", "impossible", "independent", "indestructible", "individual", "industrial", "large", "larger", "largest", "last", "least", "likely", "limited", "little", "live", "living", "local", "lonely", "loose", "lovely", "low", "lower", "lucky", "main", "major", "manly", "married", "meanest", "military", "modern", "mortal", "musical", "my", "mysterious", "national", "native", "natural", "nearest", "necessary", "negative", "nervous", "new", "next", "nice", "normal", "numerical", "old", "older", "oldest", "open", "ordinary", "original", "other", "our", "paranoid", "particular", "personal", "physical", "plural", "political", "poor", "popular", "positive", "possible", "powerful", "practical", "previous", "private", "quiet", "ready", "real", "recent", "red", "regular", "related", "religious", "remarkable", "rich", "safe", "same", "satisfied", "scientific", "second", "self-aware", "semantic", "separate", "serious", "severed", "sharp", "short", "similar", "single", "small", "smaller", "smallest", "social", "soft", "solid", "southern", "spacious", "special", "specific", "steady", "strong", "stronger", "successful", "such", "sudden", "terrible", "third", "tiny", "total", "tropical", "typical", "unhappy", "unknown", "unusual", "upper", "useful", "usual", "vacuous", "valuable", "vertical", "weak", "western", "white", "whole", "wide", "willing", "wonderful", "worried", "worse", "wrong", "young", "younger", "your"],
    "nouns": ["ability", "abstraction", "accident", "account", "acres", "act", "action", "activity", "addition", "adult", "adventure", "advertisement", "advice", "afternoon", "age", "agreement", "air", "airplane", "allowance", "alphabet", "amount", "ancestor", "angle", "animal", "answer", "ants", "anybody", "anyone", "anything", "apartment", "appearance", "apple", "appropriation", "area", "arm", "army", "arrangement", "array", "arrow", "art", "article", "artist", "astronaut", "atmosphere", "atom", "attack", "attempt", "attention", "audience", "author", "automobile", "avoidance", "baby", "bag", "bailbondsman", "balance", "ball", "balloon", "band", "bank", "bar", "barn", "bartender", "base", "baseball", "basis", "basket", "bat", "battle", "bean", "bear", "beard", "beat", "beauty", "bee", "behavior", "bell", "belongings", "belt", "bend", "bet", "bicycle", "bill", "birds", "birth", "birthday", "bit", "bite", "blanket", "blindness", "block", "blood", "blow", "board", "boat", "body", "bone", "book", "border", "bot", "bottle", "bottom", "bound", "bow", "bowl", "box", "boy", "brain", "branch", "brass", "bravery", "bread", "break", "breakfast", "breath", "breathing", "breeze", "brick", "bridge", "briefcase", "brightness", "bro", "broken", "brother", "brownie", "brush", "buffalo", "build", "building", "burn", "burst", "bus", "bush", "business", "butter", "cabal", "cabin", "caboose", "cage", "cake", "call", "calm", "camera", "camp", "canal", "cap", "capital", "captain", "car", "carbon", "card", "care", "carrier", "case", "cast", "castle", "cat", "catch", "cattle", "cause", "cave", "cell", "cent", "center", "century", "chain", "chair", "chamber", "chance", "change", "chapter", "character", "charge", "chart", "check", "cheese", "chef", "chemical", "chicken", "chief", "child", "children", "choice", "cholera", "church", "circle", "circus", "citizen", "city", "class", "classroom", "claws", "clay", "climate", "climb", "clock", "closer", "cloth", "clothes", "clothing", "cloud", "club", "coach", "coal", "coast", "coat", "coffee", "cold", "collection", "college", "colony", "color", "column", "combination", "command", "community", "company", "comparison", "compass", "composition", "compound", "condition", "congress", "consistency", "consonant", "construction", "containment", "continent", "contrast", "control", "conversation", "cookies", "coolness", "copper", "copy", "corn", "corner", "correction", "cost", "cottage", "count", "country", "couple", "courage", "course", "court", "covering", "cow", "cowboy", "crack", "cream", "creature", "crew", "crop", "cross", "crowd", "cry", "cup", "curve", "customs", "cut", "damage", "dance", "danger", "darkness", "date", "dawn", "day", "deal", "dear", "death", "decision", "deer", "definition", "degree", "dependency", "depth", "description", "desert", "design", "desk", "detail", "determination", "development", "diagram", "diameter", "dice", "difference", "difficulty", "dig", "dinner", "direction", "dirt", "disappearance", "discovery", "discus", "discussion", "disease", "dish", "disk", "distance", "divide", "division", "doctor", "dog", "doll", "dollar", "donkey", "door", "dot", "doubt", "dozen", "drawing", "dream", "dress", "drink", "drive", "driver", "drop", "drugs", "dry-cleaner", "dryness", "duck", "dude", "dullness", "dust", "duty", "eagerness", "ear", "earnings", "earth", "east", "eatery", "edge", "education", "effect", "effort", "egg", "electricity", "element", "elephant", "end", "enemy", "energy", "engine", "engineer", "enjoyment", "entrance", "environment", "equator", "equipment", "escape", "event", "everybody", "everyone", "everything", "evidence", "exactness", "examination", "example", "excellence", "exchange", "excitement", "exercise", "existence", "expansion", "experience", "experiment", "explanation", "exploration", "expression", "eye", "face", "fact", "factor", "factory", "fair", "fall", "family", "farm", "farmer", "father", "favorite", "fear", "feathers", "feature", "fee", "feet", "fellow", "fence", "field", "fierceness", "fifteen", "fifth", "fifty", "fig", "fight", "figure", "fill", "film", "finery", "finish", "fire", "fireplace", "firm", "fish", "fix", "flag", "flame", "flight", "floor", "flow", "flower", "fly", "fog", "folks", "food", "foot", "football", "force", "forest", "form", "fort", "fortnight", "forty", "fox", "frame", "freedom", "freezer", "friend", "frog", "front", "fruit", "fuel", "fun", "function", "fungus", "fur", "furniture", "future", "gain", "game", "garage", "garden", "gas", "gasoline", "gate", "gathering", "gentleness", "geranium", "giant", "gift", "glass", "globe", "gold", "goose", "government", "grade", "grain", "grandfather", "grandmother", "graph", "grass", "gravity", "gray", "ground", "group", "growth", "guard", "guess", "guide", "gulf", "gun", "habit", "hair", "half", "hall", "hand", "handle", "handmaiden", "harbor", "hardness", "hat", "hay", "health", "hearing", "heart", "heat", "height", "help", "herd", "herself", "hide", "highway", "hill", "hipster", "history", "hit", "hold", "hole", "hollow", "home", "honor", "hope", "horn", "horse", "hospital", "hour", "house", "human", "hunger", "hunt", "hunter", "hurry", "hurt", "husband", "ice", "idea", "identity", "ill", "image", "imagination", "importance", "inclusion", "income", "increase", "indication", "industry", "influence", "influencer", "information", "instance", "instant", "instrument", "interest", "interior", "iron", "island", "jar", "jet", "joinery", "journey", "joy", "judge", "jump", "jungle", "key", "kids", "kind", "kitchen", "knife", "knowledge", "label", "labor", "lack", "lake", "lamp", "land", "language", "laugh", "law", "layers", "lead", "leader", "leaf", "learning", "leather", "left", "leg", "length", "lesson", "letter", "level", "library", "life", "lift", "light", "line", "lion", "lips", "liquid", "list", "listener", "load", "location", "log", "look", "loss", "lot", "loudness", "love", "luck", "lunch", "lungs", "machine", "machinery", "madness", "magic", "magnet", "mail", "man", "manner", "manufacturer", "map", "mark", "market", "mass", "massage", "master", "material", "mathematics", "matter", "meal", "meanness", "means", "measure", "meat", "medicine", "meet", "meeting", "member", "memory", "men", "mentality", "metal", "method", "mice", "middle", "might", "mile", "milk", "mill", "mind", "mine", "minerals", "minute", "mirror", "mission", "mistake", "mix", "mixture", "model", "molecule", "moment", "money", "month", "mood", "moon", "morning", "mother", "motion", "motor", "mountain", "mouse", "mouth", "movement", "movie", "mud", "muscle", "music", "nails", "name", "nation", "nature", "neck", "needle", "needs", "neighbor", "neighborhood", "newborn", "news", "newspaper", "night", "nobody", "noise", "noodles", "noon", "north", "note", "nothing", "nothingness", "notice", "number", "object", "observation", "occurrence", "ocean", "offer", "office", "officer", "official", "oil", "operation", "opinion", "opportunity", "opposite", "orange", "orb", "orbit", "order", "organization", "origin", "outline", "owner", "oxygen", "pack", "package", "page", "pain", "paint", "pair", "palace", "pallbearer", "pan", "paper", "paragraph", "parallel", "parent", "park", "part", "particles", "parts", "party", "pass", "passage", "past", "path", "pattern", "payment", "peace", "pencil", "people", "percent", "perfection", "period", "person", "pet", "phrase", "piano", "pick", "picture", "pie", "piece", "pig", "pile", "pilot", "pine", "pinkness", "pipe", "pitcher", "place", "plain", "plan", "plane", "planet", "plant", "plastic", "plate", "plates", "play", "pleasantries", "pleasure", "pocket", "poem", "poet", "poetry", "point", "pole", "police", "policeman", "pond", "pony", "pool", "population", "porch", "port", "position", "post", "pot", "potatoes", "pound", "pour", "powder", "power", "practice", "preparation", "present", "president", "press", "pressure", "prevention", "price", "pride", "principal", "prize", "problem", "process", "produce", "product", "production", "program", "progress", "proof", "property", "propriety", "protection", "provision", "public", "pully", "pupil", "purity", "purple", "purpose", "push", "quarter", "queen", "question", "quickness", "rabbit", "race", "radio", "railroad", "rain", "raise", "ranch", "range", "rate", "rawness", "rays", "reach", "reader", "reading", "rear", "reason", "recall", "reception", "record", "reference", "region", "relationship", "repetition", "report", "representation", "requirement", "research", "respect", "rest", "result", "return", "review", "rhyme", "rhythm", "rice", "ride", "right", "ring", "rise", "river", "road", "roar", "rock", "rod", "roll", "roof", "room", "root", "rope", "rough", "round", "route", "row", "rubber", "ruby", "rule", "ruler", "rush", "saddle", "sadness", "safety", "sail", "sale", "salmon", "salt", "sand", "satellites", "satiation", "saw", "scale", "scene", "school", "science", "scientist", "score", "screen", "sea", "search", "season", "seat", "secret", "section", "seed", "selection", "sender", "sense", "sentence", "series", "service", "sets", "settlers", "shade", "shadow", "shaker", "shallowness", "shape", "share", "sheep", "sheet", "shelf", "shells", "shelter", "shine", "ship", "shirt", "shoe", "shop", "shore", "shortstop", "shot", "shoulder", "shout", "show", "shutter", "sickness", "sides", "sight", "sign", "signal", "silence", "silk", "silver", "simplicity", "sink", "sister", "sitting", "situation", "size", "skill", "skin", "sky", "slabs", "sleep", "slide", "slip", "slope", "smell", "smile", "smoke", "smoothness", "snake", "snow", "soap", "society", "soil", "soldier", "solution", "somebody", "someone", "something", "son", "song", "sort", "sound", "source", "south", "space", "speakers", "species", "speech", "speed", "spell", "spender", "spider", "spin", "spirit", "spite", "split", "sport", "spread", "spring", "square", "stage", "stairs", "stand", "standard", "star", "start", "state", "statement", "station", "stay", "steam", "steel", "steepness", "stems", "step", "stick", "stiffness", "stock", "stomach", "stone", "stoop", "stop", "store", "storm", "story", "stove", "strangeness", "stranger", "straw", "stream", "street", "strength", "stretch", "strike", "string", "strip", "structure", "struggle", "student", "subject", "substance", "success", "sugar", "suggestion", "suit", "sum", "summer", "sun", "sunlight", "supper", "supply", "support", "supposition", "surface", "surprise", "sweet", "swim", "symbol", "system", "table", "tail", "tales", "talk", "tank", "tape", "task", "taste", "tax", "tea", "teacher", "teaching", "team", "tear", "tears", "teeth", "telephone", "television", "teller", "temperature", "ten", "tent", "term", "test", "theory", "thing", "thinker", "thought", "thousand", "thread", "throw", "thumb", "tide", "tie", "tightness", "time", "tin", "tip", "title", "tobacco", "today", "tomorrow", "tone", "tongue", "tonight", "tool", "top", "topic", "touch", "tower", "town", "toy", "trace", "track", "trade", "traffic", "trail", "train", "transportation", "trap", "traveler", "tree", "triangle", "tribe", "trick", "trip", "troops", "trouble", "truck", "trunk", "truth", "try", "tube", "tune", "turn", "type", "uncle", "underline", "union", "unit", "universe", "user", "valley", "value", "vapor", "variety", "vastness", "vegetable", "vessels", "victory", "view", "village", "visit", "visitor", "voice", "void", "volume", "vote", "vowel", "voyage", "wagon", "wait", "walk", "wall", "wallflower", "warmness", "warning", "washer", "waste", "watch", "water", "wave", "way", "wealth", "weather", "weed", "week", "weight", "welcome", "west", "wetness", "whale", "wheat", "wheel", "whistle", "wild", "win", "wind", "window", "winter", "wire", "wise-men", "wish", "wolf", "wonder", "wood", "wooden", "wool", "word", "work", "worker", "world", "worry", "writer", "yard", "year", "yellowness", "yes", "yesterday", "youth", "zebra", "zero", "zipper", "zoo"]
};

function chooseOne(arr) {
    const i = Math.floor(Math.random() * arr.length);
    return arr[i];
}

function hyphenConcatChoices(first, second) {
    return [chooseOne(first), chooseOne(second)].join('-');
}

function _newLocalPart() {
    const noun = words['nouns'];
    const adjective = words['adjectives'];
    const adverbly = words['adverbs'];
    const gerunding = words['gerunds'];
    if (chooseOne([0, 1])) { return hyphenConcatChoices(adjective, noun); }
    if (chooseOne([0, 1])) { return hyphenConcatChoices(gerunding, noun); }
    return hyphenConcatChoices(gerunding, adverbly);
}

function newLocalPart() {
    let localPart;
    do {
        localPart = _newLocalPart();
    }
    // Longer than 25 chars breaks the div.
    while (localPart.length > 25);
    return localPart;
}

function setEmailAddress() {
    const emailAddress = document.getElementById("emailAddress");
    const newEmailAddress = newLocalPart() + "@nmyk.io";
    emailAddress.href = "mailto:" + newEmailAddress;
    emailAddress.innerHTML = newEmailAddress;
}
